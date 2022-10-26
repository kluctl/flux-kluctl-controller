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
<h3 id="flux.kluctl.io/v1alpha1.DurationOrNever">DurationOrNever
</h3>
<p>
(<em>Appears on:</em>
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
<code>Duration</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Never</code><br>
<em>
bool
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
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#NamespacedObjectKindReference">
github.com/fluxcd/pkg/apis/meta.NamespacedObjectKindReference
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
<code>interval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>The interval at which to reconcile the KluctlDeployment.
By default, the controller will re-deploy and validate the deployment on each reconciliation.
To override this behavior, change the DeployInterval and/or ValidateInterval values.</p>
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
When not specified, the controller uses the Interval
value to retry failures.</p>
</td>
</tr>
<tr>
<td>
<code>deployInterval</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DurationOrNever">
DurationOrNever
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployInterval specifies the interval at which to deploy the KluctlDeployment.
It defaults to the Interval value, meaning that it will re-deploy on every reconciliation.
If you set DeployInterval to a different value,</p>
</td>
</tr>
<tr>
<td>
<code>deployOnChanges</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployOnChanges will cause a re-deployment whenever the rendered resources change in the deployment.
This check is performed on every reconciliation. This means that a deployment will be triggered even before
the DeployInterval has passed in case something has changed in the rendered resources.</p>
</td>
</tr>
<tr>
<td>
<code>validateInterval</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DurationOrNever">
DurationOrNever
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ValidateInterval specifies the interval at which to validate the KluctlDeployment.
Validation is performed the same way as with &lsquo;kluctl validate -t <target>&rsquo;.
Defaults to the same value as specified in Interval.
Validate is also performed whenever a deployment is performed, independent of the value of ValidateInterval</p>
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
the context found in the kluctl target. As an alternative, specify the context to be used via &lsquo;context&rsquo;</p>
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
<code>target</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Target specifies the kluctl target to deploy. If not specified, an empty target is used that has no name and no
context. Use &lsquo;TargetName&rsquo; and &lsquo;Context&rsquo; to specify the name and context in that case.</p>
</td>
</tr>
<tr>
<td>
<code>targetNameOverride</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetNameOverride sets or overrides the target name. This is especially useful when deployment without a target.</p>
</td>
</tr>
<tr>
<td>
<code>context</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, overrides the context to be used. This will effectively make kluctl ignore the context specified
in the target.</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br>
<em>
k8s.io/apimachinery/pkg/runtime.RawExtension
</em>
</td>
<td>
<em>(Optional)</em>
<p>Args specifies dynamic target args.</p>
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
<code>validate</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Validate enables validation after deploying</p>
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
<a href="https://godoc.org/github.com/fluxcd/pkg/apis/meta#NamespacedObjectKindReference">
github.com/fluxcd/pkg/apis/meta.NamespacedObjectKindReference
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
<code>interval</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<p>The interval at which to reconcile the KluctlDeployment.
By default, the controller will re-deploy and validate the deployment on each reconciliation.
To override this behavior, change the DeployInterval and/or ValidateInterval values.</p>
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
When not specified, the controller uses the Interval
value to retry failures.</p>
</td>
</tr>
<tr>
<td>
<code>deployInterval</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DurationOrNever">
DurationOrNever
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployInterval specifies the interval at which to deploy the KluctlDeployment.
It defaults to the Interval value, meaning that it will re-deploy on every reconciliation.
If you set DeployInterval to a different value,</p>
</td>
</tr>
<tr>
<td>
<code>deployOnChanges</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeployOnChanges will cause a re-deployment whenever the rendered resources change in the deployment.
This check is performed on every reconciliation. This means that a deployment will be triggered even before
the DeployInterval has passed in case something has changed in the rendered resources.</p>
</td>
</tr>
<tr>
<td>
<code>validateInterval</code><br>
<em>
<a href="#flux.kluctl.io/v1alpha1.DurationOrNever">
DurationOrNever
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ValidateInterval specifies the interval at which to validate the KluctlDeployment.
Validation is performed the same way as with &lsquo;kluctl validate -t <target>&rsquo;.
Defaults to the same value as specified in Interval.
Validate is also performed whenever a deployment is performed, independent of the value of ValidateInterval</p>
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
the context found in the kluctl target. As an alternative, specify the context to be used via &lsquo;context&rsquo;</p>
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
<code>target</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Target specifies the kluctl target to deploy. If not specified, an empty target is used that has no name and no
context. Use &lsquo;TargetName&rsquo; and &lsquo;Context&rsquo; to specify the name and context in that case.</p>
</td>
</tr>
<tr>
<td>
<code>targetNameOverride</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetNameOverride sets or overrides the target name. This is especially useful when deployment without a target.</p>
</td>
</tr>
<tr>
<td>
<code>context</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, overrides the context to be used. This will effectively make kluctl ignore the context specified
in the target.</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br>
<em>
k8s.io/apimachinery/pkg/runtime.RawExtension
</em>
</td>
<td>
<em>(Optional)</em>
<p>Args specifies dynamic target args.</p>
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
<code>validate</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Validate enables validation after deploying</p>
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
<code>lastHandledDeployAt</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
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
<tr>
<td>
<code>rawTarget</code><br>
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
<code>rawResult</code><br>
<em>
string
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
<code>rawResult</code><br>
<em>
string
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
<h3 id="flux.kluctl.io/v1alpha1.ObjectRef">ObjectRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#flux.kluctl.io/v1alpha1.FixedImage">FixedImage</a>)
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
<code>target</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>targetNameOverride</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
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
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
