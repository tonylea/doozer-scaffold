# SPEC — Stage 3b: Helm Chart Technology

**Version:** 2.1  
**Status:** Draft  
**Related ADR:** ADR Stage 3b

---

## 0. Process Requirements

The coordinator must enforce two non-negotiable process requirements throughout all implementation work. These are hard constraints. Violation means the work is rejected, reverted, and restarted.

### 0.1 Test-Driven Development (TDD) — Mandatory

Every unit of work MUST follow strict red/green/refactor TDD:

1. **Red:** Write a failing test that specifies the expected behaviour. Run it. Confirm it fails for the expected reason.
2. **Green:** Write the minimum production code to make the test pass. Run it. Confirm it passes.
3. **Refactor:** Clean up the production code while keeping all tests green. Tests are not modified during refactor — the tests are the specification.

This cycle applies to every testable behaviour — there are no exceptions.

The project repository contains a TDD skill file at `.claude/skills/`. Lead and builder agents MUST read and follow this skill file before beginning any implementation work.

**Coordinator responsibility:** Monitor lead and builder agents for TDD compliance. If an agent writes production code before a failing test, or skips the red step, reject the work immediately. Do not allow the agent to continue and "add tests later" — revert and restart from the red step. Zero exceptions.

### 0.2 Atomic Commits — Mandatory

After TDD cycles are complete for a logical unit of work, the TDD commits (red, green, refactor) MUST be squashed into atomic commits. Each atomic commit must be:

1. **The smallest complete, meaningful change** that leaves the codebase in a working state — all tests pass, no broken imports, no dead code.
2. **Self-contained** — it does not depend on a future commit to be valid.
3. **Focused** — it does one thing.

An atomic commit always contains both the test and the production code it drove. A commit that is only tests (which would fail) is not atomic. A commit that is only production code (untested) is not atomic. The pairing of test + production code that makes it pass is the atomic unit.

The project repository contains an atomic commits skill file at `.claude/skills/`. Lead and builder agents MUST read and follow this skill file before beginning any implementation work.

**Coordinator responsibility:** The most common failure mode observed in prior stages is: TDD is followed correctly, but then the entire step is squashed into one large commit instead of multiple atomic commits. This is not acceptable. A single commit covering "add YAML definition + all its tests + acceptance tests" is too large. Monitor for this specific pattern and reject it. The agent must re-squash into atomic units.

### 0.3 Enforcement Summary

| Violation                                        | Action                                |
| ------------------------------------------------ | ------------------------------------- |
| Production code written before failing test      | Reject, revert, restart from red step |
| Tests and production code written simultaneously | Reject, revert, restart from red step |
| Tests modified during refactor step              | Reject, revert, redo refactor         |
| Single large commit covering entire step         | Reject, require atomic re-squash      |
| TDD correct but squashed into one big commit     | Reject, require atomic re-squash      |
| Agent claims "it's faster to skip TDD here"      | Reject unconditionally                |
| Agent claims "I'll add tests after"              | Reject unconditionally                |

---

## 1. Purpose

Stage 3b adds Helm as a supported technology and introduces the variant group mechanism for auto-selecting between standalone and composable modes. The variant group mechanism is applied to all existing dual-mode technologies (Terraform, Dockerfile) in the same stage — problems are solved when the mechanism is introduced, not deferred.

---

## 2. Summary of Changes

### 2.1 Schema Changes

| Change                          | Detail                                                                                                                 |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| New YAML field: `variant_group` | Optional string. Two definitions sharing the same value are variants of the same technology.                           |
| New prompt qualifier: `mode`    | Optional on prompt entries. `"composable"` means the prompt is only presented when the composable variant is resolved. |

### 2.2 Engine Changes

| Change                                      | Detail                                                                                                                                                                                         |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Variant group resolution                    | After technology selection, the engine resolves each `variant_group` to the correct definition based on whether the technology is the sole selection (standalone) or one of many (composable). |
| Prompt presents one entry per variant group | Technologies sharing a `variant_group` appear as a single entry in the prompt. Display name is the `variant_group` value.                                                                      |
| Mode-scoped prompts                         | Technology-driven prompts with `mode: "composable"` are only presented when the composable variant is resolved.                                                                                |

### 2.3 Validation Changes

| Change                             | Detail                                                                                                                                          |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Variant group validation on load   | A `variant_group` must contain exactly one `standalone: true` and one `standalone: false` definition.                                           |
| Variant group selection validation | If a `variant_group` technology is the sole selection, the standalone variant is used. If selected with others, the composable variant is used. |

### 2.4 Definition Changes

| Definition               | File                                         | Change                            |
| ------------------------ | -------------------------------------------- | --------------------------------- |
| Terraform Module         | `technologies/terraform-module.yaml`         | Add `variant_group: "Terraform"`  |
| Terraform Infrastructure | `technologies/terraform-infrastructure.yaml` | Add `variant_group: "Terraform"`  |
| Dockerfile (Image)       | `technologies/dockerfile-image.yaml`         | Add `variant_group: "Dockerfile"` |
| Dockerfile (Service)     | `technologies/dockerfile-service.yaml`       | Add `variant_group: "Dockerfile"` |

### 2.5 New Definitions

| Definition      | File                                | Mode       | Detail                                                               |
| --------------- | ----------------------------------- | ---------- | -------------------------------------------------------------------- |
| Helm Chart      | `technologies/helm-chart.yaml`      | Standalone | Chart structure at project root. `variant_group: "Helm"`.            |
| Helm Deployment | `technologies/helm-deployment.yaml` | Composable | Chart under `deploy/helm/{{.chart_name}}/`. `variant_group: "Helm"`. |

---

## 3. Variant Group Mechanism

### 3.1 Schema

The `variant_group` field is an optional string on the technology definition:

```yaml
# In helm-chart.yaml
variant_group: "Helm"
standalone: true

# In helm-deployment.yaml
variant_group: "Helm"
standalone: false
```

### 3.2 Prompt Behaviour

When loading technology definitions, the engine groups definitions by `variant_group`. For each group, the prompt presents a single entry using the `variant_group` value as the display name. Technologies without `variant_group` are presented as before (using their `name` field).

After Stage 3b, the technology prompt displays:

- Dockerfile — variant group (auto-selects Image or Service)
- Go — composable, no variant group
- Helm — variant group (auto-selects Chart or Deployment)
- PowerShell Module — standalone, no variant group
- Python — composable, no variant group
- Terraform — variant group (auto-selects Module or Infrastructure)

Six entries instead of the previous nine. The user picks technologies. The tool handles the rest.

### 3.3 Resolution Logic

After the user selects technologies:

1. For each selected technology that belongs to a `variant_group`:
   - If it is the **only** technology selected → resolve to the standalone variant.
   - If it is selected **alongside** other technologies → resolve to the composable variant.
2. The resolved definitions are passed to `Generate` as before.

This logic is universal — no technology-specific conditionals.

### 3.4 Mode-Scoped Prompts

Technology-driven prompts gain an optional `mode` field:

```yaml
prompts:
  - key: "chart_name"
    title: "Helm chart name:"
    type: "text"
    default_from: "project_name"
    mode: "composable"
```

When `mode` is omitted, the prompt is always presented. When `mode: "composable"`, the prompt is only presented if the composable variant is resolved. When `mode: "standalone"`, only if standalone.

For Helm, `chart_name` has `mode: "composable"` because the standalone variant uses `{{.ProjectName}}` directly.

### 3.5 Validation Rules

On definition load:

1. If two or more definitions share a `variant_group`, there must be exactly one with `standalone: true` and exactly one with `standalone: false`. Any other combination is an error.
2. A definition without `variant_group` that has `standalone: true` cannot be combined with other technologies (unchanged behaviour).

On config validation:

1. If a standalone-only technology (no `variant_group`, `standalone: true`) is selected alongside others, validation fails as before.
2. Variant group technologies always pass selection validation — the engine resolves to the appropriate variant.

---

## 4. Technology Definitions

### 4.1 Existing Definition Updates

The following existing definitions gain a single new field. No other changes to their content.

**`technologies/terraform-module.yaml`** — add:

```yaml
variant_group: "Terraform"
```

**`technologies/terraform-infrastructure.yaml`** — add:

```yaml
variant_group: "Terraform"
```

**`technologies/dockerfile-image.yaml`** — add:

```yaml
variant_group: "Dockerfile"
```

**`technologies/dockerfile-service.yaml`** — add:

```yaml
variant_group: "Dockerfile"
```

### 4.2 Helm Chart (Standalone Variant)

File: `technologies/helm-chart.yaml`

Standalone Helm chart project — chart structure at project root. Used when Helm is the only technology selected.

**Template Escaping:** All Helm template expressions in `content` fields use `{{"{{"}}` / `{{"}}"}}` escaping (see ADR-041). Scaffold variables like `{{.ProjectName}}` are used unescaped.

```yaml
name: "Helm Chart"
variant_group: "Helm"
standalone: true

structure:
  - path: "Chart.yaml"
    content: |
      apiVersion: v2
      name: {{.ProjectName}}
      description: A Helm chart for Kubernetes
      type: application
      version: 0.1.0
      appVersion: "0.1.0"
  - path: "values.yaml"
    content: |
      replicaCount: 1

      image:
        repository: nginx
        pullPolicy: IfNotPresent
        tag: ""

      imagePullSecrets: []
      nameOverride: ""
      fullnameOverride: ""

      serviceAccount:
        create: true
        automount: true
        annotations: {}
        name: ""

      podAnnotations: {}
      podLabels: {}

      podSecurityContext: {}

      securityContext: {}

      service:
        type: ClusterIP
        port: 80

      ingress:
        enabled: false
        className: ""
        annotations: {}
        hosts:
          - host: chart-example.local
            paths:
              - path: /
                pathType: ImplementationSpecific
        tls: []

      resources: {}

      autoscaling:
        enabled: false
        minReplicas: 1
        maxReplicas: 100
        targetCPUUtilizationPercentage: 80

      nodeSelector: {}
      tolerations: []
      affinity: {}
  - path: ".helmignore"
    content: |
      .DS_Store
      .git/
      .github/
      .gitignore
      .devcontainer/
      .editorconfig
      .gitattributes
      *.md
      LICENSE
      Makefile
      tests/
  - path: "templates/"
  - path: "templates/deployment.yaml"
    content: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
      spec:
        {{"{{"}}- if not .Values.autoscaling.enabled {{"}}"}}
        replicas: {{"{{"}} .Values.replicaCount {{"}}"}}
        {{"{{"}}- end {{"}}"}}
        selector:
          matchLabels:
            {{"{{"}}- include "{{.ProjectName}}.selectorLabels" . | nindent 6 {{"}}"}}
        template:
          metadata:
            {{"{{"}}- with .Values.podAnnotations {{"}}"}}
            annotations:
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
            {{"{{"}}- end {{"}}"}}
            labels:
              {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 8 {{"}}"}}
              {{"{{"}}- with .Values.podLabels {{"}}"}}
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
              {{"{{"}}- end {{"}}"}}
          spec:
            {{"{{"}}- with .Values.imagePullSecrets {{"}}"}}
            imagePullSecrets:
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
            {{"{{"}}- end {{"}}"}}
            serviceAccountName: {{"{{"}} include "{{.ProjectName}}.serviceAccountName" . {{"}}"}}
            securityContext:
              {{"{{"}}- toYaml .Values.podSecurityContext | nindent 8 {{"}}"}}
            containers:
              - name: {{"{{"}} .Chart.Name {{"}}"}}
                securityContext:
                  {{"{{"}}- toYaml .Values.securityContext | nindent 16 {{"}}"}}
                image: "{{"{{"}} .Values.image.repository {{"}}"}}:{{"{{"}} .Values.image.tag | default .Chart.AppVersion {{"}}"}}"
                imagePullPolicy: {{"{{"}} .Values.image.pullPolicy {{"}}"}}
                ports:
                  - name: http
                    containerPort: {{"{{"}} .Values.service.port {{"}}"}}
                    protocol: TCP
                livenessProbe:
                  httpGet:
                    path: /
                    port: http
                readinessProbe:
                  httpGet:
                    path: /
                    port: http
                resources:
                  {{"{{"}}- toYaml .Values.resources | nindent 16 {{"}}"}}
            {{"{{"}}- with .Values.nodeSelector {{"}}"}}
            nodeSelector:
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
            {{"{{"}}- end {{"}}"}}
            {{"{{"}}- with .Values.affinity {{"}}"}}
            affinity:
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
            {{"{{"}}- end {{"}}"}}
            {{"{{"}}- with .Values.tolerations {{"}}"}}
            tolerations:
              {{"{{"}}- toYaml . | nindent 8 {{"}}"}}
            {{"{{"}}- end {{"}}"}}
  - path: "templates/service.yaml"
    content: |
      apiVersion: v1
      kind: Service
      metadata:
        name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
      spec:
        type: {{"{{"}} .Values.service.type {{"}}"}}
        ports:
          - port: {{"{{"}} .Values.service.port {{"}}"}}
            targetPort: http
            protocol: TCP
            name: http
        selector:
          {{"{{"}}- include "{{.ProjectName}}.selectorLabels" . | nindent 4 {{"}}"}}
  - path: "templates/serviceaccount.yaml"
    content: |
      {{"{{"}}- if .Values.serviceAccount.create -{{"}}"}}
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: {{"{{"}} include "{{.ProjectName}}.serviceAccountName" . {{"}}"}}
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
        {{"{{"}}- with .Values.serviceAccount.annotations {{"}}"}}
        annotations:
          {{"{{"}}- toYaml . | nindent 4 {{"}}"}}
        {{"{{"}}- end {{"}}"}}
      automountServiceAccountToken: {{"{{"}} .Values.serviceAccount.automount {{"}}"}}
      {{"{{"}}- end {{"}}"}}
  - path: "templates/hpa.yaml"
    content: |
      {{"{{"}}- if .Values.autoscaling.enabled {{"}}"}}
      apiVersion: autoscaling/v2
      kind: HorizontalPodAutoscaler
      metadata:
        name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
      spec:
        scaleTargetRef:
          apiVersion: apps/v1
          kind: Deployment
          name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
        minReplicas: {{"{{"}} .Values.autoscaling.minReplicas {{"}}"}}
        maxReplicas: {{"{{"}} .Values.autoscaling.maxReplicas {{"}}"}}
        metrics:
          {{"{{"}}- if .Values.autoscaling.targetCPUUtilizationPercentage {{"}}"}}
          - type: Resource
            resource:
              name: cpu
              target:
                type: Utilization
                averageUtilization: {{"{{"}} .Values.autoscaling.targetCPUUtilizationPercentage {{"}}"}}
          {{"{{"}}- end {{"}}"}}
          {{"{{"}}- if .Values.autoscaling.targetMemoryUtilizationPercentage {{"}}"}}
          - type: Resource
            resource:
              name: memory
              target:
                type: Utilization
                averageUtilization: {{"{{"}} .Values.autoscaling.targetMemoryUtilizationPercentage {{"}}"}}
          {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}
  - path: "templates/ingress.yaml"
    content: |
      {{"{{"}}- if .Values.ingress.enabled -{{"}}"}}
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
        {{"{{"}}- with .Values.ingress.annotations {{"}}"}}
        annotations:
          {{"{{"}}- toYaml . | nindent 4 {{"}}"}}
        {{"{{"}}- end {{"}}"}}
      spec:
        {{"{{"}}- if .Values.ingress.className {{"}}"}}
        ingressClassName: {{"{{"}} .Values.ingress.className {{"}}"}}
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .Values.ingress.tls {{"}}"}}
        tls:
          {{"{{"}}- range .Values.ingress.tls {{"}}"}}
          - hosts:
              {{"{{"}}- range .hosts {{"}}"}}
              - {{"{{"}} . | quote {{"}}"}}
              {{"{{"}}- end {{"}}"}}
            secretName: {{"{{"}} .secretName {{"}}"}}
          {{"{{"}}- end {{"}}"}}
        {{"{{"}}- end {{"}}"}}
        rules:
          {{"{{"}}- range .Values.ingress.hosts {{"}}"}}
          - host: {{"{{"}} .host | quote {{"}}"}}
            http:
              paths:
                {{"{{"}}- range .paths {{"}}"}}
                - path: {{"{{"}} .path {{"}}"}}
                  pathType: {{"{{"}} .pathType {{"}}"}}
                  backend:
                    service:
                      name: {{"{{"}} include "{{.ProjectName}}.fullname" $ {{"}}"}}
                      port:
                        number: {{"{{"}} $.Values.service.port {{"}}"}}
                {{"{{"}}- end {{"}}"}}
          {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}
  - path: "templates/_helpers.tpl"
    content: |
      {{"{{"}}/*
      Expand the name of the chart.
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.name" -{{"}}"}}
      {{"{{"}}- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" {{"}}"}}
      {{"{{"}}- end {{"}}"}}

      {{"{{"}}/*
      Create a default fully qualified app name.
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.fullname" -{{"}}"}}
      {{"{{"}}- if .Values.fullnameOverride {{"}}"}}
      {{"{{"}}- .Values.fullnameOverride | trunc 63 | trimSuffix "-" {{"}}"}}
      {{"{{"}}- else {{"}}"}}
      {{"{{"}}- $name := default .Chart.Name .Values.nameOverride {{"}}"}}
      {{"{{"}}- if contains $name .Release.Name {{"}}"}}
      {{"{{"}}- .Release.Name | trunc 63 | trimSuffix "-" {{"}}"}}
      {{"{{"}}- else {{"}}"}}
      {{"{{"}}- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" {{"}}"}}
      {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}

      {{"{{"}}/*
      Create chart name and version as used by the chart label.
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.chart" -{{"}}"}}
      {{"{{"}}- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" {{"}}"}}
      {{"{{"}}- end {{"}}"}}

      {{"{{"}}/*
      Common labels
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.labels" -{{"}}"}}
      helm.sh/chart: {{"{{"}} include "{{.ProjectName}}.chart" . {{"}}"}}
      {{"{{"}} include "{{.ProjectName}}.selectorLabels" . {{"}}"}}
      {{"{{"}}- if .Chart.AppVersion {{"}}"}}
      app.kubernetes.io/version: {{"{{"}} .Chart.AppVersion | quote {{"}}"}}
      {{"{{"}}- end {{"}}"}}
      app.kubernetes.io/managed-by: {{"{{"}} .Release.Service {{"}}"}}
      {{"{{"}}- end {{"}}"}}

      {{"{{"}}/*
      Selector labels
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.selectorLabels" -{{"}}"}}
      app.kubernetes.io/name: {{"{{"}} include "{{.ProjectName}}.name" . {{"}}"}}
      app.kubernetes.io/instance: {{"{{"}} .Release.Name {{"}}"}}
      {{"{{"}}- end {{"}}"}}

      {{"{{"}}/*
      Create the name of the service account to use
      */{{"}}"}}
      {{"{{"}}- define "{{.ProjectName}}.serviceAccountName" -{{"}}"}}
      {{"{{"}}- if .Values.serviceAccount.create {{"}}"}}
      {{"{{"}}- default (include "{{.ProjectName}}.fullname" .) .Values.serviceAccount.name {{"}}"}}
      {{"{{"}}- else {{"}}"}}
      {{"{{"}}- default "default" .Values.serviceAccount.name {{"}}"}}
      {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}
  - path: "templates/NOTES.txt"
    content: |
      1. Get the application URL by running these commands:
      {{"{{"}}- if .Values.ingress.enabled {{"}}"}}
      {{"{{"}}- range $host := .Values.ingress.hosts {{"}}"}}
        {{"{{"}}- range .paths {{"}}"}}
        http{{"{{"}} if $.Values.ingress.tls {{"}}"}}s{{"{{"}} end {{"}}"}}://{{"{{"}} $host.host {{"}}"}}{{"{{"}} .path {{"}}"}}
        {{"{{"}}- end {{"}}"}}
      {{"{{"}}- end {{"}}"}}
      {{"{{"}}- else if contains "NodePort" .Values.service.type {{"}}"}}
        export NODE_PORT=$(kubectl get --namespace {{"{{"}} .Release.Namespace {{"}}"}} -o jsonpath="{.spec.ports[0].nodePort}" services {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}})
        export NODE_IP=$(kubectl get nodes --namespace {{"{{"}} .Release.Namespace {{"}}"}} -o jsonpath="{.items[0].status.addresses[0].address}")
        echo http://$NODE_IP:$NODE_PORT
      {{"{{"}}- else if contains "LoadBalancer" .Values.service.type {{"}}"}}
           NOTE: It may take a few minutes for the LoadBalancer IP to be available.
                 You can watch the status by running 'kubectl get --namespace {{"{{"}} .Release.Namespace {{"}}"}} svc -w {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}'
        export SERVICE_IP=$(kubectl get svc --namespace {{"{{"}} .Release.Namespace {{"}}"}} {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}} --template "{{"{{"}}range (index .status.loadBalancer.ingress 0){{"}}"}}{{"{{"}}. {{"}}"}}{{"{{"}}end{{"}}"}}")
        echo http://$SERVICE_IP:{{"{{"}} .Values.service.port {{"}}"}}
      {{"{{"}}- else if contains "ClusterIP" .Values.service.type {{"}}"}}
        export POD_NAME=$(kubectl get pods --namespace {{"{{"}} .Release.Namespace {{"}}"}} -l "app.kubernetes.io/name={{"{{"}} include "{{.ProjectName}}.name" . {{"}}"}},app.kubernetes.io/instance={{"{{"}} .Release.Name {{"}}"}}" -o jsonpath="{.items[0].metadata.name}")
        export CONTAINER_PORT=$(kubectl get pod --namespace {{"{{"}} .Release.Namespace {{"}}"}} $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
        echo "Visit http://127.0.0.1:8080 to use your application"
        kubectl --namespace {{"{{"}} .Release.Namespace {{"}}"}} port-forward $POD_NAME 8080:$CONTAINER_PORT
      {{"{{"}}- end {{"}}"}}
  - path: "templates/tests/test-connection.yaml"
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: "{{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}-test-connection"
        labels:
          {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
        annotations:
          "helm.sh/hook": test
      spec:
        containers:
          - name: wget
            image: busybox
            command: ['wget']
            args: ['{{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}:{{"{{"}} .Values.service.port {{"}}"}}']
        restartPolicy: Never
  - path: "charts/"
  - path: "tests/"
  - path: "tests/deployment_test.yaml"
    content: |
      suite: test deployment
      templates:
        - templates/deployment.yaml
      tests:
        - it: should create a Deployment
          asserts:
            - isKind:
                of: Deployment
        - it: should use the correct image
          set:
            image:
              repository: my-app
              tag: "1.0.0"
          asserts:
            - equal:
                path: spec.template.spec.containers[0].image
                value: "my-app:1.0.0"
        - it: should set the correct replica count
          set:
            replicaCount: 3
          asserts:
            - equal:
                path: spec.replicas
                value: 3
        - it: should not set replicas when autoscaling is enabled
          set:
            autoscaling:
              enabled: true
          asserts:
            - notExists:
                path: spec.replicas
  - path: "tests/service_test.yaml"
    content: |
      suite: test service
      templates:
        - templates/service.yaml
      tests:
        - it: should create a Service
          asserts:
            - isKind:
                of: Service
        - it: should use ClusterIP by default
          asserts:
            - equal:
                path: spec.type
                value: ClusterIP
        - it: should use the configured port
          set:
            service:
              port: 8080
              type: ClusterIP
          asserts:
            - equal:
                path: spec.ports[0].port
                value: 8080

gitignore: |
  # Helm
  *.tgz
  charts/*.tgz

devcontainer:
  features: {}
  extensions:
    - "ms-kubernetes-tools.vscode-kubernetes-tools"
  setup: |
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    helm plugin install https://github.com/helm-unittest/helm-unittest

ci:
  job_name: "helm"
  setup_steps:
    - name: "Set up Helm"
      run: |
        curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        helm plugin install https://github.com/helm-unittest/helm-unittest
  lint_steps:
    - name: "Lint chart"
      run: "helm lint ."
  test_steps:
    - name: "Unit test chart"
      run: "helm unittest ."
```

### 4.3 Helm Deployment (Composable Variant)

File: `technologies/helm-deployment.yaml`

Helm as a deployment concern within a larger project. Chart nested under `deploy/helm/{{.chart_name}}/`. Used when Helm is selected alongside other technologies.

The composable variant's YAML content mirrors the standalone variant's internal chart structure. The only differences are: paths are prefixed with `deploy/helm/{{.chart_name}}/`, template helper names use `{{.chart_name}}` instead of `{{.ProjectName}}`, and `Chart.yaml` uses `{{.chart_name}}` for the chart name. The `ci` field references `deploy/helm/{{.chart_name}}` instead of `.`.

```yaml
name: "Helm Deployment"
variant_group: "Helm"
standalone: false

prompts:
  - key: "chart_name"
    title: "Helm chart name:"
    type: "text"
    default_from: "project_name"
    mode: "composable"

structure:
  # Same internal chart structure as helm-chart.yaml, but:
  # - All paths prefixed with deploy/helm/{{.chart_name}}/
  # - Template helpers use {{.chart_name}} instead of {{.ProjectName}}
  # - Chart.yaml name field uses {{.chart_name}}
  # Full content follows the same pattern as Section 4.2.
  # Only the path prefix and chart name variable differ.
  - path: "deploy/helm/{{.chart_name}}/"
  - path: "deploy/helm/{{.chart_name}}/Chart.yaml"
    content: |
      apiVersion: v2
      name: {{.chart_name}}
      description: A Helm chart for Kubernetes
      type: application
      version: 0.1.0
      appVersion: "0.1.0"
  # ... remaining structure entries follow the identical pattern:
  # deploy/helm/{{.chart_name}}/values.yaml
  # deploy/helm/{{.chart_name}}/.helmignore
  # deploy/helm/{{.chart_name}}/templates/deployment.yaml (using {{.chart_name}} in helpers)
  # deploy/helm/{{.chart_name}}/templates/service.yaml
  # deploy/helm/{{.chart_name}}/templates/serviceaccount.yaml
  # deploy/helm/{{.chart_name}}/templates/hpa.yaml
  # deploy/helm/{{.chart_name}}/templates/ingress.yaml
  # deploy/helm/{{.chart_name}}/templates/_helpers.tpl (using {{.chart_name}} in define names)
  # deploy/helm/{{.chart_name}}/templates/NOTES.txt (using {{.chart_name}} in helpers)
  # deploy/helm/{{.chart_name}}/templates/tests/test-connection.yaml
  # deploy/helm/{{.chart_name}}/charts/
  # deploy/helm/{{.chart_name}}/tests/deployment_test.yaml
  # deploy/helm/{{.chart_name}}/tests/service_test.yaml

gitignore: |
  # Helm
  *.tgz
  charts/*.tgz

devcontainer:
  features: {}
  extensions:
    - "ms-kubernetes-tools.vscode-kubernetes-tools"
  setup: |
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    helm plugin install https://github.com/helm-unittest/helm-unittest

ci:
  job_name: "helm"
  setup_steps:
    - name: "Set up Helm"
      run: |
        curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        helm plugin install https://github.com/helm-unittest/helm-unittest
  lint_steps:
    - name: "Lint chart"
      run: "helm lint deploy/helm/{{.chart_name}}"
  test_steps:
    - name: "Unit test chart"
      run: "helm unittest deploy/helm/{{.chart_name}}"
```

The implementer must produce the full YAML file with all structure entries expanded. The content for each file mirrors Section 4.2 exactly, with the two substitutions described above (path prefix and chart name variable).

---

## 5. Acceptance Criteria

Stage 3b is complete when all of the following are true:

### Variant Group Mechanism

1. The `variant_group` schema field is supported. Two definitions sharing the same `variant_group` value are recognised as variants of the same technology.
2. On load, a `variant_group` must contain exactly one `standalone: true` and one `standalone: false` definition.
3. The prompt presents one entry per `variant_group` using the group name as the display label. "Helm" appears once, not twice. Same for "Terraform" and "Dockerfile".
4. When a `variant_group` technology is the only selection, the standalone variant is resolved.
5. When a `variant_group` technology is selected alongside others, the composable variant is resolved.
6. The `mode` qualifier on prompts works: `mode: "composable"` prompts are only presented when the composable variant is resolved.
7. Existing technologies without `variant_group` continue to work exactly as before.

### Existing Technology Migration

8. Terraform Module and Terraform Infrastructure have `variant_group: "Terraform"`. The prompt shows "Terraform" once. Selecting only Terraform produces the module layout. Selecting Terraform + Go produces the infrastructure layout.
9. Dockerfile Image and Dockerfile Service have `variant_group: "Dockerfile"`. The prompt shows "Dockerfile" once. Selecting only Dockerfile produces the image layout. Selecting Dockerfile + Go produces the service layout.
10. All existing acceptance tests for Terraform and Dockerfile pass after migration. Tests may need updating to work through variant group resolution rather than selecting specific definition keys, but the scaffolded output must be identical.

### Helm Technology

11. Both Helm definitions load, parse, and validate correctly.
12. When Helm is the sole selection, the standalone variant scaffolds the complete chart structure at the project root.
13. When Helm is selected alongside other technologies, the composable variant scaffolds the chart under `deploy/helm/{{.chart_name}}/`.
14. The `chart_name` prompt is only presented in composable mode.
15. No path conflicts exist between composable Helm and any existing composable technology.
16. The `.gitignore` correctly includes the Helm section for both variants.
17. The devcontainer includes the Kubernetes Tools extension and Helm/helm-unittest setup commands. No devcontainer features contributed.
18. CI generates `lint-helm` and `test-helm` jobs. Composable references `deploy/helm/{{.chart_name}}`. Standalone references `.`.
19. Generated Helm template files contain correct, unescaped Helm syntax. No `{{"{{"}}` literal strings in any generated file.
20. `{{.ProjectName}}` is correctly substituted in standalone output. `{{.chart_name}}` is correctly substituted in composable output.

### Quality

21. All new tests pass.
22. All existing tests continue to pass (updating test mechanics for variant group resolution is permitted; scaffolded output must not change).
23. The project's own CI pipeline passes.
24. All development followed strict TDD (Section 0.1).
25. All commits are atomic (Section 0.2).

---

## 6. Implementation Order

All steps follow strict TDD (Section 0.1) and produce atomic commits (Section 0.2).

### Step 1: Variant group schema and engine support

**Goal:** The `variant_group` field is recognised in technology definitions. The engine resolves variant groups to the correct definition based on selection context. The prompt presents one entry per variant group. Mode-scoped prompts work.

**Deliverables:**
- `variant_group` field parsed from YAML definitions.
- Load-time validation: variant groups contain exactly one standalone and one composable.
- Prompt collapses variant groups into single entries.
- Resolution logic: sole selection → standalone, multi-selection → composable.
- Mode-scoped prompt support.
- All existing tests pass — no regressions.

### Step 2: Migrate Terraform and Dockerfile to variant groups

**Goal:** Terraform and Dockerfile use variant groups. The prompt shows "Terraform" and "Dockerfile" once each. Auto-selection works. Scaffolded output is identical to before.

**Deliverables:**
- `variant_group: "Terraform"` added to both Terraform definitions.
- `variant_group: "Dockerfile"` added to both Dockerfile definitions.
- Existing tests updated to work through variant group resolution.
- Scaffolded output unchanged — verified by existing acceptance tests.

### Step 3: Helm Chart (standalone) definition

**Goal:** The standalone Helm chart definition exists and produces correct output when Helm is the only technology selected.

**Deliverables:**
- `technologies/helm-chart.yaml` with content from Section 4.2.
- Standalone scaffold produces the complete chart structure at the project root.
- Template escaping works: output files contain unescaped Helm syntax.
- `{{.ProjectName}}` substituted correctly throughout.
- CI generates correct jobs referencing `.`.

### Step 4: Helm Deployment (composable) definition

**Goal:** The composable Helm deployment definition exists and produces correct output when Helm is selected alongside other technologies.

**Deliverables:**
- `technologies/helm-deployment.yaml` with content from Section 4.3.
- Composable scaffold produces chart under `deploy/helm/{{.chart_name}}/`.
- `chart_name` prompt presented only in composable mode.
- `{{.chart_name}}` substituted correctly throughout.
- CI generates correct jobs referencing `deploy/helm/{{.chart_name}}`.
- No path conflicts with existing composable technologies.

### Step 5: Acceptance verification

**Goal:** Confirm larger deliverables work end-to-end against the acceptance criteria in Section 5. If any acceptance test fails, use TDD to fix the underlying issue.

**Deliverables:**
- Helm in isolation (standalone mode) produces correct output.
- Helm in one representative combination produces correct output, asserting only on Helm-specific output.
- Variant group auto-selection works for Helm, Terraform, and Dockerfile.
- All prior-stage tests still pass.

### Step 6: CI verification

**Goal:** All CI checks pass.

**Deliverables:**
- Push changes, verify lint + unit tests + acceptance tests pass in CI.