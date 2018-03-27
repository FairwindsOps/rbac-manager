import argparse
import kubernetes
import re
import oyaml as yaml
import logging
import os
import sys

from kubernetes.config import ConfigException
from kubernetes.client.rest import ApiException


logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')
logger = logging.getLogger(__name__)


class RBACManagerException(Exception):
    pass


class RBACManager(object):

    def __init__(self, namespace, *args, **kwargs):
        self._namespace = namespace

    def _get_config_from_file(self, file=None):
        """ Get config from file """
        self._users = yaml.load(open(file))

    def update(self, **kwargs):
        file = kwargs.get('file')
        if file is not None:
            self._get_config_from_file(file)
            self._update(self._users)
        else:
            logging.info("No config file specified")

    def controller(self):
        self._controller()

    def _controller(self):
        crds = self.k8s_client.CustomObjectsApi()
        DOMAIN = "rbac-manager.reactiveops.io"
        resource_version = ''

        while True:
            stream = kubernetes.watch.Watch().stream(crds.list_namespaced_custom_object, DOMAIN, "v1beta1", self._namespace, "rbacdefinitions")
            for event in stream:
                obj = event["object"]
                logging.debug(obj)
                new_resource_version = obj['metadata']['resourceVersion']
                operation = event['type']
                if operation in ['ADDED', 'MODIFIED'] and new_resource_version != resource_version:
                    metadata = obj.get("metadata")
                    name = metadata['name']
                    logging.info("Handling %s on %s" % (operation, name))
                    self._update(obj['rbacUsers'])
                    resource_version = new_resource_version
                elif operation in ['DELETED']:
                    metadata = obj.get("metadata")
                    name = metadata['name']
                    logging.info("Handling %s on %s" % (operation, name))
                    self._update([])
                    resource_version = ''

    def _update(self, rbac_users):
        """ Update RBAC users from a list """

        rbac_client = self.k8s_client.RbacAuthorizationV1Api()
        core_client = self.k8s_client.CoreV1Api()

        requested_role_bindings = []
        requested_cluster_role_bindings = []
        requested_service_accounts = []

        logging.debug("Finding existing Cluster Role Bindings")

        rb_response = rbac_client.list_role_binding_for_all_namespaces(label_selector="rbac-manager=reactiveops")
        existing_managed_role_bindings = rb_response.items

        logging.debug("Finding existing Role Bindings")

        crb_response = rbac_client.list_cluster_role_binding(label_selector="rbac-manager=reactiveops")
        existing_managed_cluster_role_bindings = crb_response.items

        logging.debug("Finding existing Service Accounts")

        sa_response = core_client.list_service_account_for_all_namespaces(label_selector="rbac-manager=reactiveops")
        existing_managed_service_accounts = sa_response.items

        logging.debug("Parsing provided RBAC configuration file")

        labels = {"rbac-manager": "reactiveops"}

        for rbac_user in rbac_users:
            user_kind = rbac_user.get('kind', 'User')
            subject = kubernetes.client.V1Subject(
                kind=user_kind, name=rbac_user['user'])
            escaped_user_name = re.sub('[^A-Za-z0-9]+', '-', rbac_user['user'])

            if user_kind == 'ServiceAccount':
                sa_namespace = rbac_user.get('namespace', self._namespace)
                subject = kubernetes.client.V1Subject(
                    kind=user_kind, name=escaped_user_name,
                    namespace=sa_namespace)
                metadata = kubernetes.client.V1ObjectMeta(
                    name=escaped_user_name, labels=labels, namespace=sa_namespace)
                service_account = kubernetes.client.V1ServiceAccount(
                  metadata=metadata
                )
                requested_service_accounts.append(service_account)

            if 'clusterRoleBindings' in rbac_user:
                for cluster_role_binding in rbac_user['clusterRoleBindings']:
                    role_ref = kubernetes.client.V1RoleRef(
                      api_group="rbac.authorization.k8s.io",
                      kind="ClusterRole",
                      name=cluster_role_binding['clusterRole']
                    )
                    role_binding_name = "{}-{}".format(escaped_user_name, cluster_role_binding['clusterRole'])
                    metadata = kubernetes.client.V1ObjectMeta(name=role_binding_name, labels=labels)
                    cluster_role_binding = kubernetes.client.V1ClusterRoleBinding(
                      metadata=metadata,
                      role_ref=role_ref,
                      subjects=[subject]
                    )

                    requested_cluster_role_bindings.append(cluster_role_binding)

            if 'roleBindings' in rbac_user:
                for role_binding in rbac_user['roleBindings']:
                    if 'clusterRole' in role_binding:
                        role = role_binding['clusterRole']
                        role_ref = kubernetes.client.V1RoleRef(
                          api_group="rbac.authorization.k8s.io",
                          kind="ClusterRole",
                          name=role
                        )
                    elif 'role' in role_binding:
                        role = role_binding['role']
                        role_ref = kubernetes.client.V1RoleRef(
                          api_group="rbac.authorization.k8s.io",
                          kind="Role",
                          name=role
                        )
                    else:
                        logging.error("Invalid role binding, requires 'role' or 'clusterRole' attribute")
                        break

                    if 'namespace' in role_binding:
                        namespace = role_binding['namespace']
                    else:
                        logging.error("Invalid role binding, requires 'namespace' attribute")
                        break

                    role_binding_name = "{}-{}-{}".format(re.sub('[^A-Za-z0-9]+', '-', rbac_user['user']), namespace, role)
                    metadata = kubernetes.client.V1ObjectMeta(
                      name=role_binding_name,
                      namespace=namespace,
                      labels=labels
                    )
                    role_binding = kubernetes.client.V1RoleBinding(
                      metadata=metadata,
                      role_ref=role_ref,
                      subjects=[subject]
                    )

                    requested_role_bindings.append(role_binding)

        service_accounts_to_create = requested_service_accounts[:]
        service_accounts_to_delete = existing_managed_service_accounts[:]

        logging.debug("Comparing requested ServiceAccounts with existing ones")
        for rsa in requested_service_accounts:
            for esa in existing_managed_service_accounts:
                if rsa.metadata.name == esa.metadata.name:
                    logging.debug("Existing ServiceAccount found for {}".format(rsa.metadata.name))
                    service_accounts_to_create.remove(rsa)
                    service_accounts_to_delete.remove(esa)
                    break

        cluster_role_bindings_to_create = requested_cluster_role_bindings[:]
        cluster_role_bindings_to_delete = existing_managed_cluster_role_bindings[:]

        logging.debug("Comparing requested Cluster Role Bindings with existing ones")
        for rcrb in requested_cluster_role_bindings:
            for ecrb in existing_managed_cluster_role_bindings:
                if rcrb.metadata.name == ecrb.metadata.name:
                    logging.debug("Existing Cluster Role Binding found for {}".format(rcrb.metadata.name))
                    cluster_role_bindings_to_create.remove(rcrb)
                    cluster_role_bindings_to_delete.remove(ecrb)
                    break

        role_bindings_to_create = requested_role_bindings[:]
        role_bindings_to_delete = existing_managed_role_bindings[:]

        logging.debug("Comparing requested Role Bindings with existing ones")
        for rrb in requested_role_bindings:
            for erb in existing_managed_role_bindings:
                if rrb.metadata.name == erb.metadata.name:
                    logging.debug("Existing Role Binding found for {}".format(rrb.metadata.name))
                    role_bindings_to_create.remove(rrb)
                    role_bindings_to_delete.remove(erb)
                    break

        if len(service_accounts_to_create) < 1:
            logging.info("No ServiceAccounts need to be created")
        else:
            logging.info("Creating ServiceAccounts")
            for sa in service_accounts_to_create:
                logging.info("Creating ServiceAccount: {} in {}".format(
                    sa.metadata.name, sa.metadata.namespace))
                core_client.create_namespaced_service_account(
                  namespace=sa.metadata.namespace,
                  body=sa,
                  pretty=True
                )

        if len(service_accounts_to_delete) < 1:
            logging.info("No ServiceAccounts need to be deleted")
        else:
            logging.info("Deleting ServiceAccounts")
            for sa in service_accounts_to_delete:
                logging.info("Deleting ServiceAccount: {} in {}".format(
                    sa.metadata.name, sa.metadata.namespace))
                core_client.delete_namespaced_service_account(
                  namespace=sa.metadata.namespace,
                  name=sa.metadata.name,
                  body=kubernetes.client.V1DeleteOptions(),
                  pretty=True
                )

        if len(cluster_role_bindings_to_create) < 1:
            logging.info("No Cluster Role Bindings need to be created")
        else:
            logging.info("Creating Cluster Role Bindings")
            for crb in cluster_role_bindings_to_create:
                logging.info("Creating Cluster Role Binding: {}".format(crb.metadata.name))
                rbac_client.create_cluster_role_binding(
                  body=crb,
                  pretty=True
                )

        if len(cluster_role_bindings_to_delete) < 1:
            logging.info("No Cluster Role Bindings need to be deleted")
        else:
            logging.info("Deleting Cluster Role Bindings")
            for crb in cluster_role_bindings_to_delete:
                logging.info("Deleting Cluster Role Binding: {}".format(crb.metadata.name))
                rbac_client.delete_cluster_role_binding(
                  name=crb.metadata.name,
                  body=kubernetes.client.V1DeleteOptions(),
                  pretty=True
                )

        if len(role_bindings_to_create) < 1:
            logging.info("No Role Bindings need to be created")
        else:
            logging.info("Creating Role Bindings")
            for rb in role_bindings_to_create:
                logging.info("Creating Role Binding: {} in {} namespace".format(rb.metadata.name, rb.metadata.namespace))
                try:
                    rbac_client.create_namespaced_role_binding(
                      namespace=rb.metadata.namespace,
                      body=rb,
                      pretty=True
                    )
                except ApiException, e:
                    logging.error("Error creating Role Binding: {} in {} namespace".format(rb.metadata.name, rb.metadata.namespace))
                    logging.error(e)

        if len(role_bindings_to_delete) < 1:
            logging.info("No Role Bindings need to be deleted")
        else:
            logging.info("Deleting Role Bindings")
            for rb in role_bindings_to_delete:
                logging.info("Deleting Role Binding: {} in {} namespace".format(rb.metadata.name, rb.metadata.namespace))
                try:
                    rbac_client.delete_namespaced_role_binding(
                      namespace=rb.metadata.namespace,
                      name=rb.metadata.name,
                      body=kubernetes.client.V1DeleteOptions(),
                      pretty=True
                    )
                except ApiException, e:
                    logging.error("Error deleting Role Binding: {} in {} namespace".format(rb.metadata.name, rb.metadata.namespace))
                    logging.error(e)

    def __connect(self):
        """ Connect to the Kubernetes API """

        logging.debug("Connecting to Kubernetes API")
        k8s_config_loaded = False

        try:
            logging.debug("Attempting to load incluster config")
            kubernetes.config.load_incluster_config()
            logging.debug("Successfully loaded incluster config")
        except ConfigException:
            logging.debug("Loading incluster config failed")
            logging.debug("Attempting to load kube config")
            try:
                kubernetes.config.load_kube_config()
                logging.debug("Successfully loaded kube config")
            except Exception, e:
                logging.error("Loading kube config failed, exiting")
                raise RBACManagerException("Error connecting to kubernetes: {}".format(e))

        self._k8s_client = kubernetes.client
        logging.debug(self._k8s_client)

    @property
    def k8s_client(self):
        try:
            getattr(self, "_k8s_client")
        except AttributeError:
            self.__connect()
        return self._k8s_client


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Updates RBAC cluster role bindings and role bindings.')
    parser.add_argument('--config', help='YAML configuration file to load')
    parser.add_argument('--namespace', help='Namespace for service accounts', default=os.environ.get('NAMESPACE'))
    parser.add_argument('--kubectl-auth', action='store_true', help='Use kubectl command to refresh auth (useful for GKE)')
    args = parser.parse_args()
    if args.kubectl_auth:
        os.system('kubectl get ns >/dev/null 2>&1')
    try:
        if args.config is not None:
            logging.debug("Updating RBAC from file.")
            RBACManager(args.namespace).update(file=args.config)
        else:
            if not args.namespace:
                logging.error("A specified namespace is required when running in controller-mode")
                exit(parser.print_usage())
            logging.info("Managing service accounts in namespace " + args.namespace)
            logging.debug("Starting controller.")
            RBACManager(args.namespace).controller()
    except Exception, e:
        logging.critical("Error running RBACManager: {}".format(e))
        sys.exit(1)
