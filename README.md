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
deploys:
  - name: my-app-deploy
    region: us-central1
    instance_group: my-app-instance-group
    instance_template_base: my-app-instance-template-base
    instance_template: my-app-$GITHUB_RUN_NUMBER-$GITHUB_SHA
    cloud_init: cloud-init.yml # see example dir
    labels:
      github-sha: $GITHUB_SHA
    tags:
      - my-tag123
    update_policy:
      min_ready_sec: 30

delete_instance_templates_after: false
```

### Config Reference

| Variable                                               | Description                                                                                                                                                                                                                                          |
|--------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `deploys.*.name`                                       | ***Required*** Name of the deploy                                                                                                                                                                                                                    |
| `deploys.*.project`                                    | Name of the Google Cloud project                                                                                                                                                                                                                     |
| `deploys.*.creds`                                      | Either a path or the contents of a Service Account JSON Key. Required, if not specified in Github action.                                                                                                                                            |
| `deploys.*.region`                                     | ***Required*** Region of the instance group.                                                                                                                                                                                                         |
| `deploys.*.instance_group`                             | ***Required*** Name of the instance group.                                                                                                                                                                                                           |
| `deploys.*.instance_template_base`                     | ***Required*** Instance template to be used as base.                                                                                                                                                                                                 |
| `deploys.*.instance_template`                          | ***Required*** Name of the newly created instance template.                                                                                                                                                                                          |
| `deploys.*.startup_script`                             | Path or URL to script to run when VM boots. [Read more](https://cloud.google.com/compute/docs/startupscript)                                                                                                                                         |
| `deploys.*.shutdown_script`                            | Path or URL to script to run when VM shuts down. [Read more](https://cloud.google.com/compute/docs/shutdownscript)                                                                                                                                   |
| `deploys.*.cloud_init`                                 | Path or URL to cloud-init file. [Read more](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init)                                                                                                  |
| `deploys.*.labels`                                     | A set of key/value label pairs to assign to instances.                                                                                                                                                                                               |
| `deploys.*.metadata`                                   | A set of key/value metadata pairs to make available from within instances.                                                                                                                                                                           |
| `deploys.*.tags`                                       | A list of tags to assign to instances.                                                                                                                                                                                                               |
| `deploys.*.vars`                                       | A set of additional key/value variables which will be available in either startup_script, shutdown_script or cloud_init. They take precedence over ENV vars.                                                                                         |
| `deploys.*.update_policy.type`                         | The type of update process, must be either `PROACTIVE` (default) or `OPPORTUNISTIC`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#starting_an_opportunistic_or_proactive_update) |
| `deploys.*.update_policy.replacement_method`           | What action should be used to replace instances, must be either `SUBSTITUTE` (default) or `RECREATE`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#replacement_method)           |
| `deploys.*.update_policy.minimal_action`               | Minimal action to be taken on an instance, possible values are `NONE`, `REFRESH`, `REPLACE` (default) or `RESTART`. [Read more](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#minimal_action) |
| `deploys.*.update_policy.min_ready_sec`                | Time to wait between consecutive instance updates, default is 10 seconds. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#minimum_wait_time)                                                      |
| `deploys.*.update_policy.max_surge`                    | Maximum number (or percentage, i.e. `15%`) of temporary instances to add while updating. Default is 3. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#max_surge)                                 |
| `deploys.*.update_policy.max_unavailable`              | Maximum number (or percentage, i.e. `100%`) of instances that can be offline at the same time while updating. Default is 0. [Read more](https://cloud.google.com/compute/docs/instance-groups/updating-managed-instance-groups#max_unavailable)      |
| `delete_instance_templates_after`                      | Delete old instance templates after duration, defaults to `336h` (14 days). Set to `false` to disable.                                                                                                                                               |


### Variables

Environment variables can be used in `deploy.yml`. The syntax is `$FOO` or `${FOO}` and supports
substring extraction, i.e. `${GITHUB_SHA:0:7}`. 

```
${VAR:position} - Extracts substring from $VAR at "position"
${VAR:position:length} - Extracts "length" characters of substring from $VAR at "position"
```

Environment variables or `deploys.*.vars` can be used in the `startup_script`, `shutdown_script` or `cloud_init`, see [example](example/cloud-init.yml).
The syntax is `${{FOO}}`.

Github sets a bunch of [default environment variables](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/using-environment-variables#default-environment-variables).


## Github Action Inputs

| Variable             | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| `creds`              | ***Required*** Either a path or the contents of a Service Account JSON Key. |
| `config`             | Path to config file. Default `deploy.yml` or `deploy.yaml`.                 |


### Example Usage

```
uses: mattes/gce-deploy-action@v3
with:
  creds: ${{ secrets.GOOGLE_APPLICATION_CREDENTIALS }}
  config: production.yml
```


## References

* [Managed Instance Groups](https://cloud.google.com/compute/docs/instance-groups/creating-groups-of-managed-instances)
* [Container-Optimized OS](https://cloud.google.com/container-optimized-os/)
* [cloud-init](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init)
* [startup scripts](https://cloud.google.com/compute/docs/startupscript)
* [Configuring an Instance](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#configuring_an_instance)
