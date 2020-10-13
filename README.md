# Deploy to Google Compute Engine

Github action to simplify deploys to Google Compute Engine. 

This action will...

1) Clone an existing instance template (using it as a base).
2) Update metadata config of the newly created instance template to run a startup script.
3) Tell the instance group manager to perform a rolling update with the new instance template.

## Prerequisites

Set up the following resources manually in the Cloud Console 
or use a tool like [Terraform](https://www.terraform.io).

* Create a base [instance template](https://cloud.google.com/compute/docs/instance-templates/) to be cloned by this action.
* Create a managed [instance group](https://cloud.google.com/compute/docs/instance-groups/). Please note that currently **only regional instance groups** are supported.
* Create Service Account with Roles `Compute Admin` and `Service Account User` and export a new JSON key.


## deploy.yml

By default this action expects a `deploy.yml` in the root directory of the repository.
Here is an example:

```yaml
common:
  labels:
    gitsha: ${{GITHUB_SHA}}

deploys:
  - name: my-app-deploy
    region: us-central1
    instance_group: my-app-instance-group
    instance_template_base: my-app-instance-template-base
    instance_template: my-app-${{GITHUB_RUN_NUMBER}}-${{GITHUB_SHA}}
    cloud_init: cloud-init.yml # see example dir
    labels: # will also have gitsha from common section
      version: ${{APP_VERSION}}
    tags:
      - my-tag123
    update_policy:
      min_ready_sec: 30

delete_instance_templates_after: false
```

### Config Reference

| Variable                                                | Description                                                                                                                                                                                                                                          |
|---------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `deploys.*.name`                                        | ***Required*** Name of the deploy                                                                                                                                                                                                                    |
| `deploys.*.project`                                     | Name of the Google Cloud project. Defaults to Project from Credentials.                                                                                                                                                                              |
| `deploys.*.creds`                                       | Either a path or the contents of a Service Account JSON Key. Required, if not specified in Github action.                                                                                                                                            |
| `deploys.*.region`                                      | ***Required*** Region of the instance group.                                                                                                                                                                                                         |
| `deploys.*.instance_group`                              | ***Required*** Name of the instance group.                                                                                                                                                                                                           |
| `deploys.*.instance_template_base`                      | ***Required*** Instance template to be used as base.                                                                                                                                                                                                 |
| `deploys.*.instance_template`                           | ***Required*** Name of the newly created instance template.                                                                                                                                                                                          |
| `deploys.*.startup_script`                              | Path or URL to script to run when VM boots. [Read more](https://cloud.google.com/compute/docs/startupscript)                                                                                                                                         |
| `deploys.*.shutdown_script`                             | Path or URL to script to run when VM shuts down. [Read more](https://cloud.google.com/compute/docs/shutdownscript)                                                                                                                                   |
| `deploys.*.cloud_init`                                  | Path or URL to cloud-init file. [Read more](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init)                                                                                                  |
| `deploys.*.labels`                                      | A set of key/value label pairs to assign to instances. Keys override `common.*.labels`.                                                                                                                                                              |
| `deploys.*.metadata`                                    | A set of key/value metadata pairs to make available from within instances. Keys override `common.*.metadata`.                                                                                                                                        |
| `deploys.*.tags`                                        | A list of tags to assign to instances. Tags are merged with `common.*.tags`.                                                                                                                                                                         |
| `deploys.*.vars`                                        | A set of additional key/value variables which will be available in either startup_script, shutdown_script or cloud_init. Keys override existing ENV vars and `common.*.vars`.                                                                        |
| `deploys.*.update_policy.type=PROACTIVE`                | The type of update process, must be either `PROACTIVE` (default) or `OPPORTUNISTIC`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#starting_an_opportunistic_or_proactive_update) |
| `deploys.*.update_policy.replacement_method=SUBSTITUTE` | What action should be used to replace instances, must be either `SUBSTITUTE` (default) or `RECREATE`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#replacement_method)           |
| `deploys.*.update_policy.minimal_action=REPLACE`        | Minimal action to be taken on an instance, possible values are `NONE`, `REFRESH`, `REPLACE` (default) or `RESTART`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#minimal_action) |
| `deploys.*.update_policy.min_ready_sec=10`              | Time to wait between consecutive instance updates, default is 10 seconds. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#minimum_wait_time)                                                      |
| `deploys.*.update_policy.max_surge=3`                   | Maximum number (or percentage, i.e. `15%`) of temporary instances to add while updating. Default is 3. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#max_surge)                                 |
| `deploys.*.update_policy.max_unavailable=0`             | Maximum number (or percentage, i.e. `100%`) of instances that can be offline at the same time while updating. Default is 0. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#max_unavailable)      |
| `common.project`                                        | Set default for `deploys.*.project`                                                                                                                                                                                                                  |
| `common.region`                                         | Set default for `deploys.*.region`                                                                                                                                                                                                                   |
| `common.startup_script`                                 | Set default for `deploys.*.startup_script`                                                                                                                                                                                                           |
| `common.shutdown_script`                                | Set default for `deploys.*.shutdown_script`                                                                                                                                                                                                          |
| `common.cloud_init`                                     | Set default for `deploys.*.cloud_init`                                                                                                                                                                                                               |
| `common.labels`                                         | Set default for `deploys.*.labels`                                                                                                                                                                                                                   |
| `common.metadata`                                       | Set default for `deploys.*.metadata`                                                                                                                                                                                                                 |
| `common.tags`                                           | Set default for `deploys.*.tags`                                                                                                                                                                                                                     |
| `common.vars`                                           | Set default for `deploys.*.vars`                                                                                                                                                                                                                     |
| `common.update_policy.type`                             | Set default for `deploys.*.update_policy.type`                                                                                                                                                                                                       |
| `common.update_policy.replacement_method`               | Set default for `deploys.*.update_policy.replacement_method`                                                                                                                                                                                         |
| `common.update_policy.minimal_action`                   | Set default for `deploys.*.update_policy.minimal_action`                                                                                                                                                                                             |
| `common.update_policy.min_ready_sec`                    | Set default for `deploys.*.update_policy.min_ready_sec`                                                                                                                                                                                              |
| `common.update_policy.max_surge`                        | Set default for `deploys.*.update_policy.max_surge`                                                                                                                                                                                                  |
| `common.update_policy.max_unavailable`                  | Set default for `deploys.*.update_policy.max_unavailable`                                                                                                                                                                                            |
| `delete_instance_templates_after=336h`                  | Delete old instance templates after duration, defaults to `336h` (14 days). Set to `false` to disable.                                                                                                                                               |


### Variables

Environment variables can be used in `deploy.yml`, `startup_script`, `shutdown_script` and `cloud_init` files.
The syntax is `${{FOO}}` and supports substring extraction, i.e. `${{GITHUB_SHA:0:7}}`: 

```
${{VAR:position}}        - Extracts substring from $VAR at "position"
${{VAR:position:length}} - Extracts "length" characters of substring from $VAR at "position"
```

Github sets a bunch of [default environment variables](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/using-environment-variables#default-environment-variables).


## Github Action Usage

```
uses: mattes/gce-deploy-action@v4
with:
  creds: ${{ secrets.GOOGLE_APPLICATION_CREDENTIALS }}
  config: production.yml
```

| Variable             | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| `creds`              | ***Required*** Either a path or the contents of a Service Account JSON Key. |
| `config`             | Path to config file. Default `deploy.yml` or `deploy.yaml`.                 |



## More Documentation

* [Managed Instance Groups](https://cloud.google.com/compute/docs/instance-groups/creating-groups-of-managed-instances)
* [Container-Optimized OS](https://cloud.google.com/container-optimized-os/)
* [cloud-init](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init)
* [startup scripts](https://cloud.google.com/compute/docs/startupscript)
* [Configuring an Instance](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#configuring_an_instance)
