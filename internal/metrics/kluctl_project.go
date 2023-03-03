package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	KluctlProjectControllerSubsystem = "kluctlprojects"

	DeploymentDurationKey     = "deployment_duration_seconds"
	NumberOfChangesKey        = "number_of_changes"
	NumberOfDeletedObjectsKey = "number_of_deleted_objects"
	NumberOfErrorsKey         = "number_of_errors"
	NumberOfImagesKey         = "number_of_images"
	NumberOfOrphanObjectsKey  = "number_of_orphan_objects"
	NumberOfWarningsKey       = "number_of_warnings"
	PruneDurationKey          = "prune_duration_seconds"
	ValidateDurationKey       = "validate_duration_seconds"
)

var (
	deploymentDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      DeploymentDurationKey,
		Help:      "How long a single deployment takes in seconds.",
	}, []string{"namespace", "name", "mode"})

	numberOfChanges = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlProjectControllerSubsystem,
		Name:      NumberOfChangesKey,
		Help:      "How many things has been changed by a single project.",
	}, []string{"namespace", "name"})

	numberOfDeletedObjects = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlProjectControllerSubsystem,
		Name:      NumberOfDeletedObjectsKey,
		Help:      "How many things has been deleted by a single project.",
	}, []string{"namespace", "name"})

	numberOfErrors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlProjectControllerSubsystem,
		Name:      NumberOfErrorsKey,
		Help:      "How many errors are related to a single project.",
	}, []string{"namespace", "name", "action"})

	numberOfImages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      NumberOfImagesKey,
		Help:      "Number of images of a single project.",
	}, []string{"namespace", "name"})

	numberOfOrphanObjects = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlProjectControllerSubsystem,
		Name:      NumberOfOrphanObjectsKey,
		Help:      "How many orphans are related to a single project.",
	}, []string{"namespace", "name"})

	numberOfWarnings = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: KluctlProjectControllerSubsystem,
		Name:      NumberOfWarningsKey,
		Help:      "How many warnings are related to a single project.",
	}, []string{"namespace", "name", "action"})

	pruneDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      PruneDurationKey,
		Help:      "How long a single prune takes in seconds.",
	}, []string{"namespace", "name"})

	validateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: KluctlDeploymentControllerSubsystem,
		Name:      ValidateDurationKey,
		Help:      "How long a single validate takes in seconds.",
	}, []string{"namespace", "name"})
)

func init() {
	metrics.Registry.MustRegister(deploymentDuration)
	metrics.Registry.MustRegister(numberOfChanges)
	metrics.Registry.MustRegister(numberOfDeletedObjects)
	metrics.Registry.MustRegister(numberOfErrors)
	metrics.Registry.MustRegister(numberOfImages)
	metrics.Registry.MustRegister(numberOfOrphanObjects)
	metrics.Registry.MustRegister(numberOfWarnings)
	metrics.Registry.MustRegister(pruneDuration)
	metrics.Registry.MustRegister(validateDuration)
}

func NewKluctlDeploymentDuration(namespace string, name string, mode string) prometheus.Observer {
	return deploymentDuration.WithLabelValues(namespace, name, mode)
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

func NewKluctlValidateDuration(namespace string, name string) prometheus.Observer {
	return validateDuration.WithLabelValues(namespace, name)
}