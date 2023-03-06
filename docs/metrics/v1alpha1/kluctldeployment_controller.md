<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: Metrics of the KluctlDeployment Controller
linkTitle: KluctlDeployment Controller Metrics
description: KluctlDeployment documentation
weight: 20
---
-->

# Exported Metrics References

| Metrics name                | Type      | Description                                                                           |
|-----------------------------|-----------|---------------------------------------------------------------------------------------|
| deployment_interval_seconds | Gauge     | The configured deployment interval of a single deployment.                            |
| dry_run_enabled             | Gauge     | Is dry-run enabled for a single deployment.                                           |
| last_object_status          | Gauge     | Last object status of a single deployment. Zero means failure and one means success.  |
| prune_enabled               | Gauge     | Is pruning enabled for a single deployment.                                           |
| source_spec                 | Gauge     | The configured source spec of a single deployment exported via labels.                |
