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
<a href="#flux.kluctl.io/v1alpha1.LastCommandResult">LastCommandResult</a>)
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
[]ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
[]ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
[]ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
[]ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
[]ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectSpec">KluctlProjectSpec</a>)
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
ObjectRef
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
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">KluctlDeploymentTemplateSpec</a>)
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
ObjectRef
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
<code>KluctlProjectSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectSpec">
KluctlProjectSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>KluctlDeploymentTemplateSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">
KluctlDeploymentTemplateSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlDeploymentTemplateSpec</code> are embedded into this type.)
</p>
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
<code>KluctlProjectSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectSpec">
KluctlProjectSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>KluctlDeploymentTemplateSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">
KluctlDeploymentTemplateSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlDeploymentTemplateSpec</code> are embedded into this type.)
</p>
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
<code>KluctlProjectStatus</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectStatus">
KluctlProjectStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>lastDeployResult</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.LastCommandResult">
LastCommandResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastDeployResult is the result of the last deploy command</p>
</td>
</tr>
<tr>
<td>
<code>lastPruneResult</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.LastCommandResult">
LastCommandResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastDeployResult is the result of the last prune command</p>
</td>
</tr>
<tr>
<td>
<code>lastValidateResult</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.LastValidateResult">
LastValidateResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastValidateResult is the result of the last validate command</p>
</td>
</tr>
<tr>
<td>
<code>commonLabels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>CommonLabels are the commonLabels found in the deployment project when the last deployment was done.
This is used to perform cleanup/deletion in case the KluctlDeployment project is deleted</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">KluctlDeploymentTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>, 
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTemplate">KluctlMultiDeploymentTemplate</a>)
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
<code>KluctlTimingSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlTimingSpec">
KluctlTimingSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlTimingSpec</code> are embedded into this type.)
</p>
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
the context found in the kluctl target. As an alternative, RenameContexts can be used to fix
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
This is useful when the kluctl target&rsquo;s context does not match with the
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
<code>deployMode</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployMode specifies what deploy mode should be used</p>
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
<tr>
<td>
<code>deployInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployInterval specifies the interval at which to deploy the KluctlDeployment.
This is independent of the &lsquo;Interval&rsquo; value, which only causes deployments if some deployment objects have
changed.</p>
</td>
</tr>
<tr>
<td>
<code>validateInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ValidateInterval specifies the interval at which to validate the KluctlDeployment.
Validation is performed the same way as with &lsquo;kluctl validate -t <target>&rsquo;.
Defaults to 1m.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlMultiDeployment">KluctlMultiDeployment
</h3>
<p>KluctlMultiDeployment is the Schema for the kluctlmultideployments API</p>
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
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentSpec">
KluctlMultiDeploymentSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>KluctlProjectSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectSpec">
KluctlProjectSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>KluctlTimingSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlTimingSpec">
KluctlTimingSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlTimingSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>targetPattern</code><br>
<em>
string
</em>
</td>
<td>
<p>TargetPattern is the regex pattern used to match targets</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTemplate">
KluctlMultiDeploymentTemplate
</a>
</em>
</td>
<td>
<p>Template is the object template used to create KluctlDeploymet objects</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentStatus">
KluctlMultiDeploymentStatus
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
<h3 id="flux.kluctl.io/v1alpha1.KluctlMultiDeploymentSpec">KluctlMultiDeploymentSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeployment">KluctlMultiDeployment</a>)
</p>
<p>KluctlMultiDeploymentSpec defines the desired state of KluctlMultiDeployment</p>
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
<code>KluctlProjectSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectSpec">
KluctlProjectSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>KluctlTimingSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlTimingSpec">
KluctlTimingSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlTimingSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>targetPattern</code><br>
<em>
string
</em>
</td>
<td>
<p>TargetPattern is the regex pattern used to match targets</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTemplate">
KluctlMultiDeploymentTemplate
</a>
</em>
</td>
<td>
<p>Template is the object template used to create KluctlDeploymet objects</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlMultiDeploymentStatus">KluctlMultiDeploymentStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeployment">KluctlMultiDeployment</a>)
</p>
<p>KluctlMultiDeploymentStatus defines the observed state of KluctlMultiDeployment</p>
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
<code>KluctlProjectStatus</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlProjectStatus">
KluctlProjectStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlProjectStatus</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>targetCount</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetCount is the number of targets detected</p>
</td>
</tr>
<tr>
<td>
<code>targets</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTargetStatus">
[]KluctlMultiDeploymentTargetStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Targets is the list of detected targets</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTargetStatus">KluctlMultiDeploymentTargetStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentStatus">KluctlMultiDeploymentStatus</a>)
</p>
<p>KluctlMultiDeploymentTargetStatus describes the status of a single target</p>
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
<code>name</code><br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the detected target</p>
</td>
</tr>
<tr>
<td>
<code>kluctlDeploymentName</code><br>
<em>
string
</em>
</td>
<td>
<p>KluctlDeploymentName is the name of the generated KluctlDeployment object</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlMultiDeploymentTemplate">KluctlMultiDeploymentTemplate
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentSpec">KluctlMultiDeploymentSpec</a>)
</p>
<p>KluctlMultiDeploymentTemplate is the template used to create KluctlDeployment objects</p>
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
<code>ObjectMeta</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
<p>
(Members of <code>ObjectMeta</code> are embedded into this type.)
</p>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">
KluctlDeploymentTemplateSpec
</a>
</em>
</td>
<td>
<p>Spec is the KluctlDeployment spec to be used as a template</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>KluctlTimingSpec</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.KluctlTimingSpec">
KluctlTimingSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>KluctlTimingSpec</code> are embedded into this type.)
</p>
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
the context found in the kluctl target. As an alternative, RenameContexts can be used to fix
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
This is useful when the kluctl target&rsquo;s context does not match with the
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
<code>deployMode</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployMode specifies what deploy mode should be used</p>
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
<tr>
<td>
<code>deployInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployInterval specifies the interval at which to deploy the KluctlDeployment.
This is independent of the &lsquo;Interval&rsquo; value, which only causes deployments if some deployment objects have
changed.</p>
</td>
</tr>
<tr>
<td>
<code>validateInterval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ValidateInterval specifies the interval at which to validate the KluctlDeployment.
Validation is performed the same way as with &lsquo;kluctl validate -t <target>&rsquo;.
Defaults to 1m.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlProjectSpec">KluctlProjectSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentSpec">KluctlDeploymentSpec</a>, 
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentSpec">KluctlMultiDeploymentSpec</a>)
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
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlProjectStatus">KluctlProjectStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus</a>, 
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentStatus">KluctlMultiDeploymentStatus</a>)
</p>
<p>KluctlProjectStatus defines the observed state of KluctlProjectStatus</p>
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
<code>lastAttemptedRevision</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastAttemptedRevision is the revision of the last reconciliation attempt.</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KluctlTimingSpec">KluctlTimingSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">KluctlDeploymentTemplateSpec</a>, 
<a href="#flux.kluctl.io/v1alpha1.KluctlMultiDeploymentSpec">KluctlMultiDeploymentSpec</a>)
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
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.KubeConfig">KubeConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">KluctlDeploymentTemplateSpec</a>)
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
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#SecretKeyReference">
github.com/fluxcd/pkg/apis/meta.SecretKeyReference
</a>
</em>
</td>
<td>
<p>SecretRef holds the name of a secret that contains a key with
the kubeconfig file as the value. If no key is set, the key will default
to &lsquo;value&rsquo;. The secret must be in the same namespace as
the Kustomization.
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
<h3 id="flux.kluctl.io/v1alpha1.LastCommandResult">LastCommandResult
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus</a>)
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
<code>ReconcileResultBase</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ReconcileResultBase">
ReconcileResultBase
</a>
</em>
</td>
<td>
<p>
(Members of <code>ReconcileResultBase</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>result</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">
CommandResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.LastValidateResult">LastValidateResult
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentStatus">KluctlDeploymentStatus</a>)
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
<code>ReconcileResultBase</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ReconcileResultBase">
ReconcileResultBase
</a>
</em>
</td>
<td>
<p>
(Members of <code>ReconcileResultBase</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>result</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.ValidateResult">
ValidateResult
</a>
</em>
</td>
<td>
<em>(Optional)</em>
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
<h3 id="flux.kluctl.io/v1alpha1.ObjectRef">ObjectRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.CommandResult">CommandResult</a>, 
<a href="#flux.kluctl.io/v1alpha1.DeploymentError">DeploymentError</a>, 
<a href="#flux.kluctl.io/v1alpha1.FixedImage">FixedImage</a>, 
<a href="#flux.kluctl.io/v1alpha1.ValidateResultEntry">ValidateResultEntry</a>)
</p>
<p>ObjectRef contains the information necessary to locate a resource within a cluster.</p>
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
<code>group</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>version</code><br>
<em>
string
</em>
</td>
<td>
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
</tbody>
</table>
</div>
</div>
<h3 id="flux.kluctl.io/v1alpha1.ReconcileResultBase">ReconcileResultBase
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.LastCommandResult">LastCommandResult</a>, 
<a href="#flux.kluctl.io/v1alpha1.LastValidateResult">LastValidateResult</a>)
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
<code>objectsHash</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ObjectsHash is the hash of all rendered objects</p>
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
<a href="#flux.kluctl.io/v1alpha1.KluctlDeploymentTemplateSpec">KluctlDeploymentTemplateSpec</a>)
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
<h3 id="flux.kluctl.io/v1alpha1.ValidateResult">ValidateResult
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.LastValidateResult">LastValidateResult</a>)
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
<a href="#flux.kluctl.io/v1alpha1.ObjectRef">
ObjectRef
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
