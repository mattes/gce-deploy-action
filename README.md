# Deploy to Google Compute Engine

Github action to simplify deploys to Google Compute Engine. The action will perform
a rolling update with a new Instance Template using the Instance Group Manager.


## Concept

**Managed instance groups** support stateless apps that don't depend on the 
specific state of the underlying VM instances to run. This enables features
like autoscaling and autohealing, where the managed instance group can delete
and recreate instances automatically. [Read more](https://cloud.google.com/compute/docs/instance-groups/creating-groups-of-managed-instances)

**Google Cloud Load Balancer** can use an instance group as a backend service to
distribute traffic across VM instances. [Read more](https://cloud.google.com/load-balancing/docs/backend-service#backend_services_and_autoscaled_managed_instance_groups)

Using the [Container-Optimized OS](https://cloud.google.com/container-optimized-os/)
and [cloud-init](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init), 
which can be set in the **Instance Template**, additional configuration can be applied when the instance boots.
Alternativaly [startup scripts](https://cloud.google.com/compute/docs/startupscript) can be used
to run scripts at boot time. [Read more](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#configuring_an_instance)


**tldr** This action will:

1) Clone an existing instance template (using it as base).
2) Update metadata config of the newly created instance template to run a startup script.
3) Tell the instance group manager to perform a rolling update with the new instance template.


## Prerequisites

Set up the following resources manually in the Cloud Console 
or use a tool like [Terraform](https://www.terraform.io).

* Create a base [instance template](https://cloud.google.com/compute/docs/instance-templates/) to be cloned by this action.
* Create a managed [instance group](https://cloud.google.com/compute/docs/instance-groups/). Please note that currently **only regional instance groups** are supported.
* Set up [Load Balancer](https://cloud.google.com/load-balancing/docs/) to use instance group as backend service.
* Create Service Account with Roles "Compute Admin" and "Service Account User" and export a new JSON key.


## Config

By default this action expects a `deploy.yml` in the root directory of the repository.
Environment variables (syntax `$FOO` or `${FOO}`) used in this file are replaced automatically. 
[List of default variables.](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/using-environment-variables#default-environment-variables)

| Variable                                   | Description                                                                                                                                                                                        |
|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `deploys.*.name`                           | ***Required*** Name of the deploy                                                                                                                                                                  |
| `deploys.*.project`                        | Name of the Google Cloud project                                                                                                                                                                   |
| `deploys.*.google_application_credentials` | Either a path or the contents of a Service Account JSON Key. Required, if not specified in Github action.                                                                                          |
| `deploys.*.region`                         | ***Required*** Region of the instance group.                                                                                                                                                       |
| `deploys.*.instance_group`                 | ***Required*** Name of the instance group.                                                                                                                                                         |
| `deploys.*.instance_template_base`         | ***Required*** Instance template to be used as base.                                                                                                                                               |
| `deploys.*.instance_template`              | ***Required*** Name of the newly created instance template.                                                                                                                                        |
| `deploys.*.startup_script`                 | Path to script to run when VM boots. [Read more](https://cloud.google.com/compute/docs/startupscript)                                                                                              |
| `deploys.*.shutdown_script`                | Path to script to run when VM shuts down. [Read more](https://cloud.google.com/compute/docs/shutdownscript)                                                                                        |
| `deploys.*.cloud_init`                     | Path to cloud-init file. [Read more](https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance#using_cloud-init)                                                       |
| `deploys.*.vars`                           | A set of additional key/value variables which will be available as variables (syntax `$(FOO)`) in either startup_script, shutdown_script or cloud_init. Vars take precedence over ENV vars.        |
| `deploys.*.labels`                         | A set of key/value label pairs to assign to instances.                                                                                                                                             |
| `deploys.*.metadata`                       | A set of key/value metadata pairs to make available from within instances.                                                                                                                         |
| `deploys.*.tags`                           | A list of tags to assign to instances.                                                                                                                                                             |


### Example deploy.yml

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
```


## Github Action Inputs

| Variable                         | Description                                                                 |
|----------------------------------|-----------------------------------------------------------------------------|
| `dir`                            | Working directory. Defaults to root directory of repository.                |
| `config`                         | Path to config file. Default `deploy.yml` or `deploy.yaml`.                 |
| `google_application_credentials` | Either a path or the contents of a Service Account JSON Key.                |


## Example Usage

```
uses: mattes/gce-deploy-action@master
with:
  config: production.yml
  google_application_credentials: ${{ secrets.GOOGLE_APPLICATION_CREDENTIALS }}
```

