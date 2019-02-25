package job

const (
	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// ImageScanAllJob is the name of "scanall" job in job service
	ImageScanAllJob = "IMAGE_SCAN_ALL"
	// ImageTransfer : the name of image transfer job in job service
	ImageTransfer = "IMAGE_TRANSFER"
	// ImageDelete : the name of image delete job in job service
	ImageDelete = "IMAGE_DELETE"
	// ImageReplicate : the name of image replicate job in job service
	ImageReplicate = "IMAGE_REPLICATE"
	// Replication : the name of replication job in job service
	Replication = "REPLICATION"
	// ImageGC the name of image garbage collection job in job service
	ImageGC = "IMAGE_GC"

	// JobKindGeneric : Kind of generic job
	JobKindGeneric = "Generic"
	// JobKindScheduled : Kind of scheduled job
	JobKindScheduled = "Scheduled"
	// JobKindPeriodic : Kind of periodic job
	JobKindPeriodic = "Periodic"

	// JobServiceStatusPending   : job status pending
	JobServiceStatusPending = "Pending"
	// JobServiceStatusRunning   : job status running
	JobServiceStatusRunning = "Running"
	// JobServiceStatusStopped   : job status stopped
	JobServiceStatusStopped = "Stopped"
	// JobServiceStatusCancelled : job status cancelled
	JobServiceStatusCancelled = "Cancelled"
	// JobServiceStatusError     : job status error
	JobServiceStatusError = "Error"
	// JobServiceStatusSuccess   : job status success
	JobServiceStatusSuccess = "Success"
	// JobServiceStatusScheduled : job status scheduled
	JobServiceStatusScheduled = "Scheduled"

	// JobActionStop : the action to stop the job
	JobActionStop = "stop"
)
