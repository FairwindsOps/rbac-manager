# RBAC Manager

RBAC Manager simplifies the management of role bindings in Kubernetes.

## Purpose

Ideally RBAC role bindings should be configured to allow minimal access to a cluster. That generally means specifyinc access at a namespace and user level. For example, User A may need `edit` access to namespaces 1 and 2, while User B needs `view` access to namespaces 1, 2 and 3. Just to build out that simple state, you would need to 6 different role binding YAML files. What's worse, to make User A an `admin` of namespace 1, we could not just update the existing role binding. Instead, that role binding would have to be deleted and a new replacement one created.

With RBAC Manager, we can represent the state described above in a single YAML file:

```
- user: a@example.com
  roleBindings:
    - clusterRole: admin
      namespace: namespace-1
    - clusterRole: edit
      namespace: namespace-2
- user: b@example.com
  roleBindings:
    - clusterRole: view
      namespace: namespace-1
    - clusterRole: view
      namespace: namespace-2
    - clusterRole: view
      namespace: namespace-3
```

Running RBAC Manager with the above configuration will create all these requested role bindings for you. Importantly, it will compare the requested state with the existing state and only make changes necessary to reach the requested state. It uses Kubernetes labels to track which role bindings it manages. Any role bindings with that label that are not described in the configuration it is using will be removed, and any role bindings described in that configuration that do not exist will be added.

## Usage

At it's core, this is a Python script that runs with YAML configuration. There are many ways to use this, we'll cover 3 of them here:

### As a Python Script

Potentially most straightforward, this requires Python 2.7 or newer to get started. From there, we'll need to install a few dependencies:

```
pip install -r requirements.txt
```

Once you have a YAML config file that you're happy with (see the examples/config directory for examples), you can run it with this command:

```
python manage_rbac.py --config path/to/your/config.yaml
```

As you might expect, this will run in your current Kubernetes context. If you don't have Kube config set up locally, you'll need to do that. If you happen to be using GKE, the Python Kubernetes client often has problems renewing credentials. Running any `kubectl` command should renew credentials and allow this script to work.

### As a Kubernetes Job

Also quite straightforward, you can apply the YAML from the `example/k8s` directory of this repository to run RBAC Manager within your cluster. In this case, you'll want to add you're RBAC Manager configuration in the ConfigMap (`example/k8s/2-config.yaml`).

Once the ConfigMap represents the RBAC state you want to achieve, you can run the job with a simple command:

```
kubectl apply -f example/k8s
```

Once the job has completed, you can clean things up by removing the namespace it creates with this command:

```
kubectl delete namespace rbac-manager
```

### As part of a CI Workflow

Ideally RBAC manager will be used in a CI workflow. We provide a `-ci` image with every release that includes some helpful dependencies for continuous integration. There is a working example of what this could look like in `examples/ci`.

## License
Apache License 2.0