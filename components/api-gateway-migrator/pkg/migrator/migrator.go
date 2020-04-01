package migrator

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	apiruleapi "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	newapi "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	RetriesCount        uint
	DelayBetweenSteps   uint
	DelayBetweenRetries uint
}

type Migrator struct {
	DelaySecBetweenSteps uint
	oldAPI               *oldapi.Api //This instance is never mutated!
	newApiName           string
	alreadyMigrated      bool
	failure              error
	k8sClient            *K8sClient
}

func New(cl *K8sClient) *Migrator {
	//increase randomness
	rand.Seed(time.Now().UnixNano())

	return &Migrator{
		DelaySecBetweenSteps: 2, //TODO: Make it a parameter
		k8sClient:            cl,
	}
}

func (m *Migrator) MigrateOldApi(oldApi *oldapi.Api) (*MigrationResult, error) {
	m.oldAPI = oldApi

	log.Infof("Migrating: %s", oldApi.Name)
	return m.
		findOrCreateTemporaryNewApi().
		disableOldApi().
		wait().
		enableNewApi().
		wait().
		verifyNewApi().
		markOldApiMigrated().
		result()
}

func (m *Migrator) wait() *Migrator {
	if m.skip() {
		return m
	}

	time.Sleep(time.Duration(m.DelaySecBetweenSteps) * time.Second)
	return m
}

func (m *Migrator) result() (*MigrationResult, error) {

	if m.failure != nil {
		return nil, m.failure
	}

	return &MigrationResult{
		NewApiName: m.newApiName,
	}, nil
}

func (m *Migrator) skip() bool {
	return m.alreadyMigrated || m.failure != nil
}

/*
* Designed to resume previous, perhaps broken, migration.
 */
func (m *Migrator) findOrCreateTemporaryNewApi() *Migrator {

	temporaryApiRule, err := m.k8sClient.findTemporaryApiRule(m.oldAPI.Name)
	if err != nil {
		m.failure = err
		return m
	}

	if temporaryApiRule != nil {
		m.newApiName = temporaryApiRule.Name
		if isMigrated(temporaryApiRule) {
			m.alreadyMigrated = true
			log.Infof("api %s is already migrated to APIRule: %s", m.oldAPI.Name, temporaryApiRule.Name)
			return m
		}

		log.Infof("Found existing temporary APIRule: %s", temporaryApiRule.Name)
		return m
	}

	apiRuleName := generateApiRuleName(m.oldAPI)
	apiRuleHost := generateTemporaryHost(m.oldAPI)

	temporaryApiRule = translateToApiRule(m.oldAPI)
	temporaryApiRule.Name = apiRuleName
	temporaryApiRule.Spec.Service.Host = &apiRuleHost

	setNewApiLabel(temporaryApiRule, "targetHost", m.oldAPI.Spec.Hostname)

	log.Infof("creating a temporary APIRule: %s", temporaryApiRule.Name)
	err = m.createApiRule(temporaryApiRule)
	if err != nil {
		m.failure = err
	}
	m.newApiName = temporaryApiRule.Name

	return m
}

// change the hostname of old api to some temporary value
func (m *Migrator) disableOldApi() *Migrator {
	if m.skip() {
		return m
	}

	if m.oldAPI.Labels != nil && m.oldAPI.Labels["migration/status"] != "" {
		return m
	}

	log.Infof("disabling old api: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)

	oldApiHost := m.oldAPI.Spec.Hostname
	tempOldApiHost := generateTemporaryHost(m.oldAPI)

	var updateFunc oldApiUpdateFunc = func(oldApi *oldapi.Api) {
		oldApi.Spec.Hostname = tempOldApiHost
		setOldApiLabel(oldApi, "migration/host", oldApiHost)
		setOldApiLabel(oldApi, "migration/status", "disabled")
	}

	if err := m.updateOldApi(updateFunc); err != nil {
		m.failure = err
		return m
	}

	log.Infof("successfully disabled old api: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)
	return m
}

// change the hostname of new apirule to the proper one
func (m *Migrator) enableNewApi() *Migrator {
	if m.skip() {
		return m
	}

	log.Infof("enabling apirule for api: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)
	host := m.oldAPI.Spec.Hostname

	var updateFunc newApiUpdateFunc = func(newApi *newapi.APIRule) {
		newApi.Spec.Service.Host = &host
		setNewApiLabel(newApi, "migratedFrom", m.oldAPI.Name)
	}

	if err := m.updateNewApiRule(updateFunc); err != nil {
		m.failure = err
		// TODO: check if error is returned when nothing changed in apirule (host was already changed to proper value)
		return m
	}
	log.Infof("successfully enabled apirule for api: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)
	return m
}

// set label on an old api meaning that the object was migrated
func (m *Migrator) markOldApiMigrated() *Migrator {
	if m.skip() {
		return m
	}

	if m.oldAPI.Labels != nil && m.oldAPI.Labels["migration/status"] == "migrated" {
		return m
	}

	log.Infof("marking old api as migrated: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)

	var updateFunc oldApiUpdateFunc = func(oldApi *oldapi.Api) {
		oldApi.Labels["migration/status"] = "migrated"
	}

	if err := m.updateOldApi(updateFunc); err != nil {
		m.failure = err
		return m
	}

	log.Infof("successfully marked old api as migrated: %s/%s", m.oldAPI.Namespace, m.oldAPI.Name)
	return m
}

func (m *Migrator) verifyNewApi() *Migrator {
	if m.skip() {
		return m
	}

	newApi, err := m.k8sClient.findTemporaryApiRule(m.oldAPI.Name)
	if err != nil {
		m.failure = err
		return m
	}
	if newApi.Status.APIRuleStatus.Code != newapi.StatusOK {
		m.failure = errors.New(fmt.Sprintf("Invalid ApiRule status: %s", newApi.Status.APIRuleStatus.Code))
		return m
	}
	if newApi.Status.VirtualServiceStatus.Code != newapi.StatusOK {
		m.failure = errors.New(fmt.Sprintf("Invalid VirtualService status: %s", newApi.Status.VirtualServiceStatus.Code))
		return m
	}
	if newApi.Status.AccessRuleStatus.Code != newapi.StatusOK {
		m.failure = errors.New(fmt.Sprintf("Invalid AccessRule status: %s", newApi.Status.AccessRuleStatus.Code))
	}
	return m
}

func generateApiRuleName(oldApi *oldapi.Api) string {
	return oldApi.Name + "-" + generateRandomString(6)
}

func generateTemporaryHost(oldApi *oldapi.Api) string {
	return generateRandomString(6) + "-" + oldApi.Spec.Hostname
}

func (m *Migrator) createApiRule(apirule *apiruleapi.APIRule) error {
	return m.k8sClient.create(apirule)
}

func (m *Migrator) updateOldApi(update oldApiUpdateFunc) error {
	objKey := client.ObjectKey{
		Name:      m.oldAPI.Name,
		Namespace: m.oldAPI.Namespace,
	}

	return m.k8sClient.getAndUpdateOldApi(objKey, update)
}

func (m *Migrator) updateNewApiRule(update newApiUpdateFunc) error {

	objKey := client.ObjectKey{
		Name:      m.newApiName,
		Namespace: m.oldAPI.Namespace,
	}

	return m.k8sClient.getAndUpdateNewApi(objKey, update)
}

type MigrationResult struct {
	NewApiName string
	Failures   []error
}

func generateRandomString(n uint) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//TODO: Find a nice function to handle both apis.
func setNewApiLabel(object *newapi.APIRule, key, value string) {
	if object.Labels == nil {
		object.Labels = map[string]string{}
	}
	object.Labels[key] = value
}

func setOldApiLabel(object *oldapi.Api, key, value string) {
	if object.Labels == nil {
		object.Labels = map[string]string{}
	}
	object.Labels[key] = value
}

func isMigrated(obj *newapi.APIRule) bool {
	if obj.Labels == nil {
		return false
	}

	return obj.Labels["migratedFrom"] != ""
}
