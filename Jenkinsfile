// ========================================
// Jenkins Pipeline for Proxy Server
// ========================================
// 
// 项目：Proxy Server Web Manager
// 作者：Dong hua
// 
// 功能特性：
// - 前后端一体化构建
// - Docker 镜像构建与推送
// - Harbor ImagePullSecret 自动创建
// - Kubernetes 多端口部署
// - 健康检查
// ========================================

pipeline {
    agent any

    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 30, unit: 'MINUTES')
        timestamps()
    }

    environment {
        // Docker 仓库配置
        DOCKER_REGISTRY = '192.168.1.8:8083'
        IMAGE_NAME = 'proxy-server/proxy-server'
        IMAGE_FULL = "${DOCKER_REGISTRY}/${IMAGE_NAME}"
        
        // 构建配置
        TAG = "${env.BUILD_NUMBER}"
        
        // Kubernetes 配置
        K8S_NAMESPACE = 'default'
        K8S_APP_NAME = 'proxy-server'
        K8S_DEPLOYMENT_NAME = 'proxy-server'
        
        // 容器端口配置
        CONTAINER_PORT_API = '8000'          // 后端API端口
        CONTAINER_PORT_FRONTEND = '3000'     // 前端端口
        CONTAINER_PORT_SOCKS5 = '10808'      // SOCKS5代理端口
        CONTAINER_PORT_HTTP = '10809'        // HTTP代理端口
        CONTAINER_PORT_MIXED = '10810'       // 混合端口
        
        // NodePort配置 (K8s默认范围: 30000-32767)
        NODEPORT_API = '30800'
        NODEPORT_FRONTEND = '30300'
        NODEPORT_SOCKS5 = '30808'
        NODEPORT_HTTP = '30809'
        NODEPORT_MIXED = '30810'
        
        // Jenkins Credentials ID
        HARBOR_CREDENTIAL_ID = 'Harbor'
        KUBECONFIG_CREDENTIAL_ID = 'kubeconfig'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
                echo "Checkout completed: ${env.GIT_COMMIT}"
            }
        }

        stage('Build Docker Image') {
            steps {
                sh """
                    echo "Building Docker image..."
                    docker build -t ${IMAGE_NAME}:${TAG} .
                    docker build -t ${IMAGE_NAME}:latest .
                    echo "Docker image built: ${IMAGE_NAME}:${TAG}"
                """
            }
        }

        stage('Push to Harbor') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: "${HARBOR_CREDENTIAL_ID}",
                    usernameVariable: 'HARBOR_USER',
                    passwordVariable: 'HARBOR_PASS'
                )]) {
                    sh """
                        echo "Logging in to Harbor..."
                        echo "${HARBOR_PASS}" | docker login ${DOCKER_REGISTRY} -u ${HARBOR_USER} --password-stdin
                        
                        echo "Tagging and pushing images..."
                        docker tag ${IMAGE_NAME}:${TAG} ${IMAGE_FULL}:${TAG}
                        docker tag ${IMAGE_NAME}:latest ${IMAGE_FULL}:latest
                        docker push ${IMAGE_FULL}:${TAG}
                        docker push ${IMAGE_FULL}:latest
                        
                        echo "Images pushed successfully"
                    """
                }
            }
        }

        stage('Create ImagePullSecret') {
            steps {
                withCredentials([file(
                    credentialsId: "${KUBECONFIG_CREDENTIAL_ID}",
                    variable: 'KUBECONFIG'
                )]) {
                    withCredentials([usernamePassword(
                        credentialsId: "${HARBOR_CREDENTIAL_ID}",
                        usernameVariable: 'HARBOR_USER',
                        passwordVariable: 'HARBOR_PASS'
                    )]) {
                        sh """
                            export KUBECONFIG=\$KUBECONFIG
                            
                            echo "Creating Harbor ImagePullSecret..."
                            kubectl delete secret harbor-reg -n ${K8S_NAMESPACE} --ignore-not-found=true
                            kubectl create secret docker-registry harbor-reg \\
                              --docker-server=${DOCKER_REGISTRY} \\
                              --docker-username=\${HARBOR_USER} \\
                              --docker-password=\${HARBOR_PASS} \\
                              --docker-email=admin@example.com \\
                              -n ${K8S_NAMESPACE}
                            
                            echo "ImagePullSecret created successfully"
                        """
                    }
                }
            }
        }

        stage('Clean Old Resources') {
            steps {
                withCredentials([file(
                    credentialsId: "${KUBECONFIG_CREDENTIAL_ID}",
                    variable: 'KUBECONFIG'
                )]) {
                    sh """
                        export KUBECONFIG=\$KUBECONFIG
                        
                        echo "Deleting old resources..."
                        kubectl delete deployment ${K8S_DEPLOYMENT_NAME} -n ${K8S_NAMESPACE} --ignore-not-found=true --force --grace-period=0
                        kubectl delete service ${K8S_DEPLOYMENT_NAME} -n ${K8S_NAMESPACE} --ignore-not-found=true
                        
                        echo "Waiting for resources to be deleted..."
                        sleep 5
                    """
                }
            }
        }

        stage('Deploy to K8s') {
            steps {
                withCredentials([file(
                    credentialsId: "${KUBECONFIG_CREDENTIAL_ID}",
                    variable: 'KUBECONFIG'
                )]) {
                    sh """
                        export KUBECONFIG=\$KUBECONFIG
                        
                        echo "Creating new resources..."
                        kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${K8S_DEPLOYMENT_NAME}
  namespace: ${K8S_NAMESPACE}
  labels:
    app: ${K8S_APP_NAME}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${K8S_APP_NAME}
  template:
    metadata:
      labels:
        app: ${K8S_APP_NAME}
    spec:
      imagePullSecrets:
      - name: harbor-reg
      containers:
      - name: ${K8S_APP_NAME}
        image: ${IMAGE_FULL}:${TAG}
        imagePullPolicy: Always
        ports:
        - containerPort: ${CONTAINER_PORT_API}
          name: api
        - containerPort: ${CONTAINER_PORT_FRONTEND}
          name: frontend
        - containerPort: ${CONTAINER_PORT_SOCKS5}
          name: socks5
        - containerPort: ${CONTAINER_PORT_HTTP}
          name: http-proxy
        - containerPort: ${CONTAINER_PORT_MIXED}
          name: mixed
        env:
        - name: GIN_MODE
          value: "release"
        - name: PORT
          value: "${CONTAINER_PORT_API}"
        resources:
          requests:
            memory: "512Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /api/status
            port: ${CONTAINER_PORT_API}
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/status
            port: ${CONTAINER_PORT_API}
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: ${K8S_DEPLOYMENT_NAME}
  namespace: ${K8S_NAMESPACE}
spec:
  type: NodePort
  selector:
    app: ${K8S_APP_NAME}
  ports:
  - name: api
    port: ${CONTAINER_PORT_API}
    targetPort: ${CONTAINER_PORT_API}
    nodePort: ${NODEPORT_API}
    protocol: TCP
  - name: frontend
    port: ${CONTAINER_PORT_FRONTEND}
    targetPort: ${CONTAINER_PORT_FRONTEND}
    nodePort: ${NODEPORT_FRONTEND}
    protocol: TCP
  - name: socks5
    port: ${CONTAINER_PORT_SOCKS5}
    targetPort: ${CONTAINER_PORT_SOCKS5}
    nodePort: ${NODEPORT_SOCKS5}
    protocol: TCP
  - name: http-proxy
    port: ${CONTAINER_PORT_HTTP}
    targetPort: ${CONTAINER_PORT_HTTP}
    nodePort: ${NODEPORT_HTTP}
    protocol: TCP
  - name: mixed
    port: ${CONTAINER_PORT_MIXED}
    targetPort: ${CONTAINER_PORT_MIXED}
    nodePort: ${NODEPORT_MIXED}
    protocol: TCP
EOF

                        echo "Waiting for deployment to complete..."
                        kubectl rollout status deployment/${K8S_DEPLOYMENT_NAME} -n ${K8S_NAMESPACE} --timeout=300s
                    """
                }
            }
        }

        stage('Health Check') {
            steps {
                withCredentials([file(
                    credentialsId: "${KUBECONFIG_CREDENTIAL_ID}",
                    variable: 'KUBECONFIG'
                )]) {
                    sh """
                        export KUBECONFIG=\$KUBECONFIG
                        
                        echo "Checking pod status..."
                        for i in {1..30}; do
                            POD_STATUS=\$(kubectl get pods -n ${K8S_NAMESPACE} -l app=${K8S_APP_NAME} -o jsonpath="{.items[0].status.phase}" 2>/dev/null)
                            if [ "\$POD_STATUS" = "Running" ]; then
                                echo "Pod is running"
                                exit 0
                            fi
                            echo "Waiting for Pod... (\$i/30), status: \$POD_STATUS"
                            sleep 2
                        done
                        
                        echo "Health check timeout"
                        exit 1
                    """
                }
            }
        }
    }

    post {
        success {
            echo """
            ========================================
            BUILD SUCCESS!
            ========================================
            Image: ${IMAGE_FULL}:${TAG}
            Namespace: ${K8S_NAMESPACE}
            Deployment: ${K8S_DEPLOYMENT_NAME}
            
            端口映射:
            - API:       NodePort ${NODEPORT_API} -> Container ${CONTAINER_PORT_API}
            - Frontend:  NodePort ${NODEPORT_FRONTEND} -> Container ${CONTAINER_PORT_FRONTEND}
            - SOCKS5:    NodePort ${NODEPORT_SOCKS5} -> Container ${CONTAINER_PORT_SOCKS5}
            - HTTP:      NodePort ${NODEPORT_HTTP} -> Container ${CONTAINER_PORT_HTTP}
            - Mixed:     NodePort ${NODEPORT_MIXED} -> Container ${CONTAINER_PORT_MIXED}
            ========================================
            """
        }
        failure {
            echo """
            ========================================
            BUILD FAILED!
            ========================================
            Please check the logs for details.
            ========================================
            """
        }
    }
}
