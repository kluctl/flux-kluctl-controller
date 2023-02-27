package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Metrics subsystem and all keys used by the kluctldeployment controller.
const (
	KluctlDeploymentControllerSubsystem = "kluctldeployments"

	DeploymentDurationKey     = "deployment_duration_seconds"
	DeploymentIntervalKey     = "deployment_interval_seconds"
	DryRunEnabledKey          = "dry_run_enabled"
	LastObjectStatusKey       = "last_object_status"
	NumberOfChangesKey        = "number_of_changes"
	NumberOfDeletedObjectsKey = "number_of_deleted_objects"
	NumberOfErrorsKey         = "number_of_errors"
	NumberOfImagesKey         = "number_of_images"
	NumberOfOrphanObjectsKey  = "number_of_orphan_objects"
	NumberOfWarningsKey       = "number_of_warnings"
	PruneDurationKey          = "prune_duration_seconds"
	PruneEnabledKey           = "prune_enabled"
	SourceSpecKey             = "source_spec"
	ValidateDurationKey       = "validate_duration_seconds"
)

var (
	deploymentDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      DeploymentDurationKey,
		Help:      "How long a single deployment takes in seconds.",
	}, []string{"namespace", "name", "mode"})

	deploymentInterval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      DeploymentIntervalKey,
		Help:      "The configured deployment interval of a single deployment.",
	}, []string{"namespace", "name"})

	dryRunEnabled = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      DryRunEnabledKey,
		Help:      "Is dry-run enabled for a single deployment.",
	}, []string{"namespace", "name"})

	lastObjectStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      LastObjectStatusKey,
		Help:      "Last object status of a single deployment.",
	}, []string{"namespace", "name"})

	numberOfChanges = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfChangesKey,
		Help:      "How many things has been changed by a single deployment.",
	}, []string{"namespace", "name"})

	numberOfDeletedObjects = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfDeletedObjectsKey,
		Help:      "How many things has been deleted by a single deployment.",
	}, []string{"namespace", "name"})

	numberOfErrors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfErrorsKey,
		Help:      "How many errors are related to a single deployment.",
	}, []string{"namespace", "name", "action"})

	numberOfImages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfImagesKey,
		Help:      "Number of images of a single deployment.",
	}, []string{"namespace", "name"})

	numberOfOrphanObjects = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfOrphanObjectsKey,
		Help:      "How many orphans are related to a single deployment.",
	}, []string{"namespace", "name"})

	numberOfWarnings = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfWarningsKey,
		Help:      "How many warnings are related to a single deployment.",
	}, []string{"namespace", "name", "action"})

	pruneDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      PruneDurationKey,
		Help:      "How long a single prune takes in seconds.",
	}, []string{"namespace", "name"})

	pruneEnabled = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      PruneEnabledKey,
		Help:      "Is pruning enabled for a single deployment.",
	}, []string{"namespace", "name"})

	sourceSpec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      SourceSpecKey,
		Help:      "The configured source spec of a single deployment.",
	}, []string{"namespace", "name", "url", "path", "ref"})

	validateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      ValidateDurationKey,
		Help:      "How long a single validate takes in seconds.",
	}, []string{"namespace", "name"})
)

func init() {
	metrics.Registry.MustRegister(deploymentDuration)
	metrics.Registry.MustRegister(deploymentInterval)
	metrics.Registry.MustRegister(dryRunEnabled)
	metrics.Registry.MustRegister(lastObjectStatus)
	metrics.Registry.MustRegister(numberOfChanges)
	metrics.Registry.MustRegister(numberOfDeletedObjects)
	metrics.Registry.MustRegister(numberOfErrors)
	metrics.Registry.MustRegister(numberOfImages)
	metrics.Registry.MustRegister(numberOfOrphanObjects)
	metrics.Registry.MustRegister(numberOfWarnings)
	metrics.Registry.MustRegister(pruneDuration)
	metrics.Registry.MustRegister(pruneEnabled)
	metrics.Registry.MustRegister(sourceSpec)
	metrics.Registry.MustRegister(validateDuration)
}

func NewKluctlDeploymentDuration(namespace string, name string, mode string) prometheus.Observer {
	return deploymentDuration.WithLabelValues(namespace, name, mode)
}

func NewKluctlDeploymentInterval(namespace string, name string) prometheus.Gauge {
	return dryRunEnabled.WithLabelValues(namespace, name)
}

func NewKluctlDryRunEnabled(namespace string, name string) prometheus.Gauge {
	return dryRunEnabled.WithLabelValues(namespace, name)
}

func NewKluctlLastObjectStatus(namespace string, name string) prometheus.Gauge {
	return lastObjectStatus.WithLabelValues(namespace, name)
}

func NewKluctlNumberOfChanges(namespace string, name string) prometheus.Gauge {
	return numberOfChanges.WithLabelValues(namespace, name)
}

func NewKluctlNumberOfDeletedObjects(namespace string, name string) prometheus.Gauge {
	return numberOfDeletedObjects.WithLabelValues(namespace, name)
}

func NewKluctlNumberOfErrors(namespace string, name string, action string) prometheus.Gauge {
	return numberOfErrors.WithLabelValues(namespace, name, action)
}

func NewKluctlNumberOfImages(namespace string, name string) prometheus.Gauge {
	return numberOfImages.WithLabelValues(namespace, name)
}

func NewKluctlNumberOfOrphanObjects(namespace string, name string) prometheus.Gauge {
	return numberOfOrphanObjects.WithLabelValues(namespace, name)
}

func NewKluctlNumberOfWarnings(namespace string, name string, action string) prometheus.Gauge {
	return numberOfWarnings.WithLabelValues(namespace, name, action)
}

func NewKluctlPruneDuration(namespace string, name string) prometheus.Observer {
	return pruneDuration.WithLabelValues(namespace, name)
}

func NewKluctlPruneEnabled(namespace string, name string) prometheus.Gauge {
	return pruneEnabled.WithLabelValues(namespace, name)
}

func NewKluctlSourceSpec(namespace string, name string, url string, path string, ref string) prometheus.Gauge {
	return sourceSpec.WithLabelValues(namespace, name, url, path, ref)
}

func NewKluctlValidateDuration(namespace string, name string) prometheus.Observer {
	return validateDuration.WithLabelValues(namespace, name)
}
