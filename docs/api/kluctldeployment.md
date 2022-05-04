<h1>KluctlDeployment API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#flux.kluctl.io%2fv1alpha1">flux.kluctl.io/v1alpha1</a>
</li>
</ul>
<h2 id="flux.kluctl.io/v1alpha1">flux.kluctl.io/v1alpha1</h2>
<p>Package v1alpha1 contains API Schema definitions for the flux.kluctl.io v1alpha1 API group.</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="flux.kluctl.io/v1alpha1.Change">Change
</h3>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>jsonPath</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>unifiedDiff</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.CommandResult">CommandResult
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.ReconcileAttempt">ReconcileAttempt</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>newObjects</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
[]ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>changedObjects</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
[]ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>hookObjects</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
[]ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>orphanObjects</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
[]ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deletedObjects</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
[]ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>errors</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">
[]DeploymentError
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>warnings</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">
[]DeploymentError
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>seenImages</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.FixedImage">
[]FixedImage
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.CrossNamespaceSourceReference">CrossNamespaceSourceReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>)
</p>
<p>CrossNamespaceSourceReference contains enough information to let you locate the
typed Kubernetes resource object at cluster level.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>API version of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
<em>
string
</em>
</td>
<td>
<p>Kind of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br>
<em>
string
</em>
</td>
<td>
<p>Name of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the referent, defaults to the namespace of the Kubernetes resource object that contains the reference.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.DeploymentError">DeploymentError
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">CommandResult</a>, 
<a href="#flux.kluctl.io/v1alpha1.ValidateResult">ValidateResult</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ref</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>error</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.FixedImage">FixedImage
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">CommandResult</a>, 
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>resultImage</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deployedImage</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>registryImage</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>object</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deployment</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>container</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>versionFilter</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deployTags</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deploymentDir</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.InvolvedRepo">InvolvedRepo
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus</a>)
</p>
<p>InvolvedRepo represents a git repository and all involved refs</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code><br>
<em>
string
</em>
</td>
<td>
<p>URL is the url of the involved git repository</p>
</td>
</tr>
<tr>
<td>
<code>patterns</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.InvolvedRepoPattern">
[]InvolvedRepoPattern
</a>
</em>
</td>
<td>
<p>Patterns is a list of pattern+refs combinations</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.InvolvedRepoPattern">InvolvedRepoPattern
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.InvolvedRepo">InvolvedRepo</a>)
</p>
<p>InvolvedRepoPattern represents a ref pattern and the found refs</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>pattern</code><br>
<em>
string
</em>
</td>
<td>
<p>Pattern is a regex to filter refs</p>
</td>
</tr>
<tr>
<td>
<code>refs</code><br>
<em>
map[string]string
</em>
</td>
<td>
<p>Refs is the filtered list of refs</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlDeployment">KluctlDeployment
</h3>
<p>KluctlDeployment is the Schema for the kluctldeployments API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">
KluctlDeploymentSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>dependsOn</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#NamespacedObjectReference">
[]github.com/fluxcd/pkg/apis/meta.NamespacedObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DependsOn may contain a meta.NamespacedObjectReference slice
with references to resources that must be ready before this
kluctl project can be deployed.</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>The interval at which to reconcile the KluctlDeployment.</p>
</td>
</tr>
<tr>
<td>
<code>retryInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The interval at which to retry a previously failed reconciliation.
When not specified, the controller uses the KluctlDeploymentSpec.Interval
value to retry failures.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Path to the directory containing the .kluctl.yaml file, or the
Defaults to &lsquo;None&rsquo;, which translates to the root path of the SourceRef.</p>
</td>
</tr>
<tr>
<td>
<code>sourceRef</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.CrossNamespaceSourceReference">
CrossNamespaceSourceReference
</a>
</em>
</td>
<td>
<p>Reference of the source where the kluctl project is.
The authentication secrets from the source are also used to authenticate
dependent git repositories which are cloned while deploying the kluctl project.</p>
</td>
</tr>
<tr>
<td>
<code>registrySecrets</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#LocalObjectReference">
[]github.com/fluxcd/pkg/apis/meta.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RegistrySecrets is a list of secret references to be used for image registry authentication.
The secrets must either have &ldquo;.dockerconfigjson&rdquo; included or &ldquo;registry&rdquo;, &ldquo;username&rdquo; and &ldquo;password&rdquo;.
Additionally, &ldquo;caFile&rdquo; and &ldquo;insecure&rdquo; can be specified.</p>
</td>
</tr>
<tr>
<td>
<code>suspend</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>This flag tells the controller to suspend subsequent kluctl executions,
it does not apply to already started executions. Defaults to false.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Timeout for all operations.
Defaults to &lsquo;Interval&rsquo; duration.</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br>
<em>
string
</em>
</td>
<td>
<p>Target specifies the kluctl target to deploy</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the Kubernetes service account to use while deploying.
If not specified, the default service account is used.</p>
</td>
</tr>
<tr>
<td>
<code>kubeConfig</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KubeConfig">
KubeConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The KubeConfig for deploying to the target cluster.
Specifies the kubeconfig to be used when invoking kluctl. Contexts in this kubeconfig must match
the cluster config found in the kluctl project. As alternative, RenameContexts can be used to fix
non-matching context names.</p>
</td>
</tr>
<tr>
<td>
<code>renameContexts</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.RenameContext">
[]RenameContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RenameContexts specifies a list of context rename operations.
This is useful when the kluctl project&rsquo;s cluster configs specify contexts that do not match with the
contexts found in the kubeconfig while deploying. This is the case when using kubeconfigs generated from
service accounts, in which case the context name is always &ldquo;default&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Args specifies dynamic target args.
Only arguments defined by &lsquo;dynamicArgs&rsquo; of the target are allowed.</p>
</td>
</tr>
<tr>
<td>
<code>updateImages</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>UpdateImages instructs kluctl to update dynamic images.
Equivalent to using &lsquo;-u&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>images</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.FixedImage">
[]FixedImage
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Images contains a list of fixed image overrides.
Equivalent to using &lsquo;&ndash;fixed-images-file&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>dryRun</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DryRun instructs kluctl to run everything in dry-run mode.
Equivalent to using &lsquo;&ndash;dry-run&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>noWait</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>NoWait instructs kluctl to not wait for any resources to become ready, including hooks.
Equivalent to using &lsquo;&ndash;no-wait&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>forceApply</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceApply instructs kluctl to force-apply in case of SSA conflicts.
Equivalent to using &lsquo;&ndash;force-apply&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>replaceOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ReplaceOnError instructs kluctl to replace resources on error.
Equivalent to using &lsquo;&ndash;replace-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>forceReplaceOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceReplaceOnError instructs kluctl to force-replace resources in case a normal replace fails.
Equivalent to using &lsquo;&ndash;force-replace-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>abortOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceReplaceOnError instructs kluctl to abort deployments immediately when something fails.
Equivalent to using &lsquo;&ndash;abort-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>includeTags</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IncludeTags instructs kluctl to only include deployments with given tags.
Equivalent to using &lsquo;&ndash;include-tag&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>excludeTags</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeTags instructs kluctl to exclude deployments with given tags.
Equivalent to using &lsquo;&ndash;exclude-tag&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>includeDeploymentDirs</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IncludeDeploymentDirs instructs kluctl to only include deployments with the given dir.
Equivalent to using &lsquo;&ndash;include-deployment-dir&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>excludeDeploymentDirs</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeDeploymentDirs instructs kluctl to exclude deployments with the given dir.
Equivalent to using &lsquo;&ndash;exclude-deployment-dir&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>prune</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prune enables pruning after deploying.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">
KluctlDeploymentStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeployment">KluctlDeployment</a>)
</p>
<p>KluctlDeploymentSpec defines the desired state of KluctlDeployment</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>dependsOn</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#NamespacedObjectReference">
[]github.com/fluxcd/pkg/apis/meta.NamespacedObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DependsOn may contain a meta.NamespacedObjectReference slice
with references to resources that must be ready before this
kluctl project can be deployed.</p>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>The interval at which to reconcile the KluctlDeployment.</p>
</td>
</tr>
<tr>
<td>
<code>retryInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The interval at which to retry a previously failed reconciliation.
When not specified, the controller uses the KluctlDeploymentSpec.Interval
value to retry failures.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Path to the directory containing the .kluctl.yaml file, or the
Defaults to &lsquo;None&rsquo;, which translates to the root path of the SourceRef.</p>
</td>
</tr>
<tr>
<td>
<code>sourceRef</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.CrossNamespaceSourceReference">
CrossNamespaceSourceReference
</a>
</em>
</td>
<td>
<p>Reference of the source where the kluctl project is.
The authentication secrets from the source are also used to authenticate
dependent git repositories which are cloned while deploying the kluctl project.</p>
</td>
</tr>
<tr>
<td>
<code>registrySecrets</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#LocalObjectReference">
[]github.com/fluxcd/pkg/apis/meta.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RegistrySecrets is a list of secret references to be used for image registry authentication.
The secrets must either have &ldquo;.dockerconfigjson&rdquo; included or &ldquo;registry&rdquo;, &ldquo;username&rdquo; and &ldquo;password&rdquo;.
Additionally, &ldquo;caFile&rdquo; and &ldquo;insecure&rdquo; can be specified.</p>
</td>
</tr>
<tr>
<td>
<code>suspend</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>This flag tells the controller to suspend subsequent kluctl executions,
it does not apply to already started executions. Defaults to false.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Timeout for all operations.
Defaults to &lsquo;Interval&rsquo; duration.</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br>
<em>
string
</em>
</td>
<td>
<p>Target specifies the kluctl target to deploy</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the Kubernetes service account to use while deploying.
If not specified, the default service account is used.</p>
</td>
</tr>
<tr>
<td>
<code>kubeConfig</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KubeConfig">
KubeConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The KubeConfig for deploying to the target cluster.
Specifies the kubeconfig to be used when invoking kluctl. Contexts in this kubeconfig must match
the cluster config found in the kluctl project. As alternative, RenameContexts can be used to fix
non-matching context names.</p>
</td>
</tr>
<tr>
<td>
<code>renameContexts</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.RenameContext">
[]RenameContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RenameContexts specifies a list of context rename operations.
This is useful when the kluctl project&rsquo;s cluster configs specify contexts that do not match with the
contexts found in the kubeconfig while deploying. This is the case when using kubeconfigs generated from
service accounts, in which case the context name is always &ldquo;default&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Args specifies dynamic target args.
Only arguments defined by &lsquo;dynamicArgs&rsquo; of the target are allowed.</p>
</td>
</tr>
<tr>
<td>
<code>updateImages</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>UpdateImages instructs kluctl to update dynamic images.
Equivalent to using &lsquo;-u&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>images</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.FixedImage">
[]FixedImage
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Images contains a list of fixed image overrides.
Equivalent to using &lsquo;&ndash;fixed-images-file&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>dryRun</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DryRun instructs kluctl to run everything in dry-run mode.
Equivalent to using &lsquo;&ndash;dry-run&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>noWait</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>NoWait instructs kluctl to not wait for any resources to become ready, including hooks.
Equivalent to using &lsquo;&ndash;no-wait&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>forceApply</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceApply instructs kluctl to force-apply in case of SSA conflicts.
Equivalent to using &lsquo;&ndash;force-apply&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>replaceOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ReplaceOnError instructs kluctl to replace resources on error.
Equivalent to using &lsquo;&ndash;replace-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>forceReplaceOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceReplaceOnError instructs kluctl to force-replace resources in case a normal replace fails.
Equivalent to using &lsquo;&ndash;force-replace-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>abortOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ForceReplaceOnError instructs kluctl to abort deployments immediately when something fails.
Equivalent to using &lsquo;&ndash;abort-on-error&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>includeTags</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IncludeTags instructs kluctl to only include deployments with given tags.
Equivalent to using &lsquo;&ndash;include-tag&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>excludeTags</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeTags instructs kluctl to exclude deployments with given tags.
Equivalent to using &lsquo;&ndash;exclude-tag&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>includeDeploymentDirs</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>IncludeDeploymentDirs instructs kluctl to only include deployments with the given dir.
Equivalent to using &lsquo;&ndash;include-deployment-dir&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>excludeDeploymentDirs</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExcludeDeploymentDirs instructs kluctl to exclude deployments with the given dir.
Equivalent to using &lsquo;&ndash;exclude-deployment-dir&rsquo; when calling kluctl.</p>
</td>
</tr>
<tr>
<td>
<code>prune</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prune enables pruning after deploying.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeployment">KluctlDeployment</a>)
</p>
<p>KluctlDeploymentStatus defines the observed state of KluctlDeployment</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ReconcileRequestStatus</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#ReconcileRequestStatus">
github.com/fluxcd/pkg/apis/meta.ReconcileRequestStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>ReconcileRequestStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code><br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>ObservedGeneration is the last reconciled generation.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>lastForceReconcileHash</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastForceReconcileHash contains a hash of all values from the spec that must cause a forced
reconcile.</p>
</td>
</tr>
<tr>
<td>
<code>involvedRepos</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.InvolvedRepo">
[]InvolvedRepo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>InvolvedRepos is a list of repositories and refs involved with this kluctl project</p>
</td>
</tr>
<tr>
<td>
<code>lastAttemptedReconcile</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ReconcileAttempt">
ReconcileAttempt
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The last attempted reconcile.</p>
</td>
</tr>
<tr>
<td>
<code>lastSuccessfulReconcile</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ReconcileAttempt">
ReconcileAttempt
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The last successfully reconcile attempt.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KubeConfig">KubeConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>)
</p>
<p>KubeConfig references a Kubernetes secret that contains a kubeconfig file.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretRef</code><br>
<em>
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#LocalObjectReference">
github.com/fluxcd/pkg/apis/meta.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecretRef holds the name to a secret that contains a &lsquo;value&rsquo; key with
the kubeconfig file as the value. It must be in the same namespace as
the KluctlDeployment.
It is recommended that the kubeconfig is self-contained, and the secret
is regularly updated if credentials such as a cloud-access-token expire.
Cloud specific <code>cmd-path</code> auth helpers will not function without adding
binaries and credentials to the Pod that is responsible for reconciling
the KluctlDeployment.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.ReconcileAttempt">ReconcileAttempt
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus</a>)
</p>
<p>ReconcileAttempt describes an attempt to reconcile</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>time</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>AttemptedAt is the time when the attempt was performed</p>
</td>
</tr>
<tr>
<td>
<code>revision</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Revision is the source revision. Please note that kluctl projects have
dependent git repositories which are not considered in the source revision</p>
</td>
</tr>
<tr>
<td>
<code>targetName</code><br>
<em>
string
</em>
</td>
<td>
<p>TargetName is the name of the target</p>
</td>
</tr>
<tr>
<td>
<code>targetHash</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetHash is the hash of the target configuration</p>
</td>
</tr>
<tr>
<td>
<code>deployResult</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">
CommandResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployResult is the command result of the deploy command</p>
</td>
</tr>
<tr>
<td>
<code>pruneResult</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">
CommandResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PruneResult is the command result of the prune command</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.RenameContext">RenameContext
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>)
</p>
<p>RenameContext specifies a single rename of a context</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>oldContext</code><br>
<em>
string
</em>
</td>
<td>
<p>OldContext is the name of the context to be renamed</p>
</td>
</tr>
<tr>
<td>
<code>newContext</code><br>
<em>
string
</em>
</td>
<td>
<p>NewContext is the new name of the context</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.ResourceRef">ResourceRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">CommandResult</a>, 
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">DeploymentError</a>, 
<a href="#flux.kluctl.io/v1alpha1.FixedImage">FixedImage</a>, 
<a href="#flux.kluctl.io/v1alpha1.ValidateResultEntry">ValidateResultEntry</a>)
</p>
<p>ResourceRef contains the information necessary to locate a resource within a cluster.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>id</code><br>
<em>
string
</em>
</td>
<td>
<p>ID is the string representation of the Kubernetes resource object&rsquo;s metadata,
in the format &lsquo;<namespace><em><name></em><group>_<kind>&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>v</code><br>
<em>
string
</em>
</td>
<td>
<p>Version is the API version of the Kubernetes resource object&rsquo;s kind.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.ValidateResult">ValidateResult
</h3>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ready</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>warnings</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">
[]DeploymentError
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>errors</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">
[]DeploymentError
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>results</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ValidateResultEntry">
[]ValidateResultEntry
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.ValidateResultEntry">ValidateResultEntry
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.ValidateResult">ValidateResult</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ref</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ResourceRef">
ResourceRef
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>annotation</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>