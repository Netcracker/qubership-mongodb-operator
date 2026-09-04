# Copyright 2024-2025 NetCracker Technology Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import time, sys
import kubernetes
import urllib3
from kubernetes.stream import stream
from kubernetes import client, config
from robot.api import logger
try:
    from openshift.dynamic import ResourceInstance
except ImportError:
    ResourceInstance = None
    logger.warn("openshift module not available; OpenShift features disabled")


class KubernetesClient(object):
    SA_NAMESPACE_PATH = '/var/run/secrets/kubernetes.io/serviceaccount/namespace'
    def __init__(self, cluster=None, token=None):
        self.namespace = open(self.SA_NAMESPACE_PATH).read()
        urllib3.disable_warnings()
        if cluster is None:
            try:
                config.load_incluster_config()
                self._api_client = client.ApiClient()
            except config.ConfigException as e:
                print("Can't load incluster kubernetes config. This script is intended to use inside of kubernetes")
        else:
            conf = client.Configuration()
            conf.host = cluster
            conf.verify_ssl = False
            conf.api_key = {"authorization": "Bearer " + token}
            self._api_client = client.ApiClient(conf)

        self._apps_v1_api = client.AppsV1Api(self._api_client)
        self._core_v1_api = client.CoreV1Api(self._api_client)

    def get_deployment_scale(self, name):
        return self._apps_v1_api.read_namespaced_deployment_scale(
            name=name, namespace=self.namespace)

    def set_deployment_scale(self, name, namespace, scale):
        self._apps_v1_api.replace_namespaced_deployment_scale(name=name,
                                                              namespace=self.namespace,
                                                              body=scale,
                                                              pretty='true')

    def get_deployments(self) -> client.V1DeploymentList:
        return self._apps_v1_api.list_namespaced_deployment(namespace=self.namespace)

    def get_var_on_pod(self, var, label):
        pod_list = self.get_pods(labels=label)
        pod: client.V1PodList
        for pod in pod_list.items:
                envs = pod.spec.containers[0].env
                for env in envs:
                    if env.name == var:
                        return env.value
        return None

    def get_external_mounts(self, service, label):
        pod_list = self.get_pods(labels=label)
        pod: client.V1PodList
        for pod in pod_list.items:
            if pod.metadata.labels.get('name') == service:
                mounts = pod.spec.containers[0].volume_mounts
                for mount in mounts:
                    sys.__stdout__.write('MOUNT %s\n' % mount.name)
                    if mount.name == 'external-backup-storage':
                        return True
        return None

    def get_deployments_for_service(self, service, label_name='clusterName'):
        """
        This method gets all found (active/inactive) deployments name for given service.
        *Args:*\n
            _namespace_ (str) - OpenShift project name;\n
            _service_ (str) - service name;\n
        *Return:*\n
            list(str) - list of deployments name for given service
        """
        all_deployments_in_project = self.get_deployments()
        deployments = []
        for deployment in all_deployments_in_project.items:
            if deployment.spec.template.metadata.labels.get(label_name, '') == service:
                deployments.append(deployment.metadata.name)
        return deployments

    def get_pods(self, labels=None):
        return self._core_v1_api.list_namespaced_pod(namespace=self.namespace, label_selector=labels)

    def delete_pod(self, name):
        body = kubernetes.client.V1DeleteOptions()
        self._core_v1_api.delete_namespaced_pod(name, body)

    def get_pods_for_service(self, namespace, service, label_name='name'):
        """
        This method gets all found pods for given service.
        *Args:*\n
            _namespace_ (str) - OpenShift project name;\n
            _service_ (str) - service name;\n
        *Return:*\n
            V1PodList - list of pods
        *Example:*\n
            | Get Pods For Service | kafka-service | kafka |
        """
        all_pods_in_project = self._core_v1_api.list_namespaced_pod(namespace=namespace)
        pods = []
        for pod in all_pods_in_project.items:
            # sys.__stdout__.write('Got POD %s\n' % pod)
            if pod.metadata.labels.get(label_name, '') == service:
                pods.append(pod.metadata.name)
        return pods

    def get_pod_for_service(self, namespace, service, label_name='name'):
        """
        This method gets first found pod name for given service. Pod connects with service by labels.
        So we need to get selector from service and get pod with the same label.
        *Args:*\n
            _namespace_ (str) - OpenShift project name;\n
            _service_ (str) - service name to get pod;\n
        *Return:*\n
            str - pod name for given service
        *Example:*\n
            | Get Pod For Service | kafka-service | kafka |
        """
        pods = self.get_pods_for_service(namespace, service, label_name)
        for pod in pods:
            return pod
        return None

    def execute_command_in_pod(self, pod_name, namespace, command):
        sys.__stdout__.write('Got command %s\n' % command)
        exec_command = [
            '/bin/bash',
            '-c',
            command]
        resp = stream(self._core_v1_api.connect_get_namespaced_pod_exec,
                      name=pod_name,
                      namespace=namespace,
                      command=exec_command,
                      stderr=True, stdin=False,
                      stdout=True, tty=False)
        print("Response: " + resp)
        resp_type = type(resp)
        sys.__stdout__.write('Got type %s\n' % resp_type)
        sys.__stdout__.write('Got resp %s\n' % resp)
        return resp

    def list_namespaced_statefulsets(self):
        return self._apps_v1_api.list_namespaced_stateful_set(self.namespace)

    def scale_statefulset(self, name_statefulset, state: int):
        all_statefulsets = self.list_namespaced_statefulsets()
        for stateful in all_statefulsets.items:
            if stateful.metadata.name in name_statefulset[0]:
                stateful.spec.replicas = state
                resp = self._apps_v1_api.patch_namespaced_stateful_set(stateful.metadata.name, self.namespace, stateful)
                return resp.spec.replicas
        return None

    def get_side_for_pod(self, namespace, service, left_pattern='left', right_pattern='right'):
        namespaced_pods = self._core_v1_api.list_namespaced_pod(namespace=namespace)
        pods_node = []
        for pod in namespaced_pods.items:
            if service in pod.metadata.name:
                pods_node.append(pod.spec.node_name)
        left_pods = 0
        right_pods = 0
        left_nodes_pattern = self.convert_labels_to_list(left_pattern)
        right_nodes_pattern = self.convert_labels_to_list(right_pattern)
        for pod in pods_node:
            if any(label in pod for label in left_nodes_pattern):
                left_pods += 1
            elif any(label in pod for label in right_nodes_pattern):
                right_pods += 1
        return left_pods, right_pods

    def convert_labels_to_list(self, labels):
        list_labels = labels.split(",")
        return list_labels

    @staticmethod
    def _get_environments_for_container(containers, container_name):
        environments = None
        for container in containers:
            if container.name == container_name:
                environments = container.env
        return environments

    @staticmethod
    def _get_env_variables(dicts: list, params: list, ignore_reference=True):
        if not dicts:
            return None
        result = {}
        for dictionary in dicts:
            if dictionary.name in params:
                if not ignore_reference:
                    result[dictionary.name] = dictionary.value if dictionary.value is not None else ""
                elif dictionary.value is not None:
                    result[dictionary.name] = dictionary.value
        return result

    def _get_environment_variables_for_container(self,
                                                 entity,
                                                 container_name: str,
                                                 variable_names: list) -> dict:
        environments = self._get_environments_for_container(entity.spec.template.spec.containers, container_name)
        return self._get_env_variables(environments, variable_names)

    def get_environment_variables_for_deployment_entity_container(self,
                                                                  name: str,
                                                                  namespace: str,
                                                                  container_name: str,
                                                                  variable_names: list) -> dict:
        entity = self._apps_v1_api.read_namespaced_deployment(name, namespace)
        return self._get_environment_variables_for_container(entity, container_name, variable_names)

    # def encode_url(self, url):
    #   '''
    #   Encodes the given url that would (probably) contain characters outside the ASCII set into a valid ASCII format
    #   `Returns: string`
    #   '''
    #   return urllib.parse.quote(url)

    def get_daemon_set(self, name: str, namespace: str) -> ResourceInstance:
        """
        Returns daemon set by name in specified project/namespace.

        :param namespace: namespace to find daemon set
        :param name: name to find daemon set
        :return: a found daemon set in the namespace

        Example:
        | Get Daemon Set | node-exporter | prometheus-operator |
        """
        ret = self.get_daemon_sets(namespace)
        items = [item for item in ret if name == item.metadata.name]
        return items[0]

    def get_image(self, resource, container_name):
        """
        Returns image from resource configuration by container name in specified project/namespace.
        """
        if len(resource.spec.template.spec.containers) > 1 and container_name is not None:
            for container in resource.spec.template.spec.containers:
                if container.name == container_name:
                    return container.image
        elif len(resource.spec.template.spec.containers) == 1:
            return resource.spec.template.spec.containers[0].image
        return None


    def get_deployment_entity(self, name: str, namespace: str):
        """Returns `deployment/deployment config` configuration.
         Example:
         | Get Deployment Entity | elasticsearch-1 | elasticsearch-service |
        """
        return self._apps_v1_api.read_namespaced_deployment(name, namespace)
    def get_deployment_entity_names_by_service_name(self, service_name: str, namespace: str) -> list:
        """Returns list of `deployment`/`deployment config` names by given Kubernetes service name and `project`/`namespace`.
        There is no direct mapping between `deployment entity` and `service`. Supposed that deployment entity watches the
        same kubernetes `pods` as Kubernetes service. So the `deployment entity` matches to the `service` by the
        transitivity property.

        Method raises an Exception if `Service` or `namespace` is not found.

        Example:
        | Get Deployment Entity Names By Service Name | monitoring-collector | postgres-service |
        """
        selector = self.get_service_selector(service_name, namespace)
        return self.get_deployment_entity_names_by_selector(namespace, selector)

    def get_stateful_set(self, name: str, namespace: str):
        """Returns `Stateful Set` configuration
        `Stateful Set` is found by name and namespace/project.

        Method raises an Exception if Stateful Set is not found.

        Example:
        | Get Stateful Set | cassandra0 | cassandra |
        """
        return self._apps_v1_api.read_namespaced_stateful_set(name, namespace)

    def get_resource_image(self, resource_type: str, resource_name: str, namespace: str, resource_container_name=None):
        """
        Identifies the resource type and return image for the specified resource by the name of the resource and container in the specified project/namespace.
        """
        if resource_type == 'deployment':
            deployment = self.get_deployment_entity(resource_name, namespace)
            return self.get_image(deployment, resource_container_name)
        elif resource_type == 'statefulset':
            stateful_set = self.get_stateful_set(resource_name, namespace)
            return self.get_image(stateful_set, resource_container_name)
        else:
            raise Exception(f'The type [{resource_type}] is not supported yet.')

    def get_config_map(self, name: str, namespace: str):
        """
        Returns config map by name in specified project/namespace.

        Example:
        | Get Config Map | elasticsearch-config-map | elasticsearch |
        """
        return self._core_v1_api.read_namespaced_config_map(name, namespace)

    def get_dd_images_from_config_map(self, config_map_name, namespace):
        config_map = self.get_config_map(config_map_name, namespace)
        config_map_yaml = (config_map.to_dict())
        cm = config_map_yaml["data"]["dd_images"]
        if cm:
            return cm
        else:
            return None
