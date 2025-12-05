<p>Packages:</p>
<ul>
<li>
<a href="#landscape.config.gardener.cloud%2fv1alpha1">landscape.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="landscape.config.gardener.cloud/v1alpha1">landscape.config.gardener.cloud/v1alpha1</h2>
<p>
</p>
Resource Types:
<ul></ul>
<h3 id="landscape.config.gardener.cloud/v1alpha1.GitRepository">GitRepository
</h3>
<p>
(<em>Appears on:</em>
<a href="#landscape.config.gardener.cloud/v1alpha1.LandscapeKitConfiguration">LandscapeKitConfiguration</a>)
</p>
<p>
<p>GitRepository contains information the Git repository containing landscape deployments and configurations.</p>
</p>
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
<code>url</code></br>
<em>
string
</em>
</td>
<td>
<p>URL specifies the Git repository URL, it can be an HTTP/S or SSH address.</p>
</td>
</tr>
<tr>
<td>
<code>ref</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.GitRepositoryRef">
GitRepositoryRef
</a>
</em>
</td>
<td>
<p>Reference specifies the Git reference to resolve and monitor for
changes, defaults to the &lsquo;master&rsquo; branch.</p>
</td>
</tr>
<tr>
<td>
<code>paths</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.PathConfiguration">
PathConfiguration
</a>
</em>
</td>
<td>
<p>Paths specifies the path configuration within the Git repository.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.GitRepositoryRef">GitRepositoryRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#landscape.config.gardener.cloud/v1alpha1.GitRepository">GitRepository</a>)
</p>
<p>
<p>GitRepositoryRef specifies the Git reference to resolve and checkout.</p>
</p>
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
<code>branch</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Branch to check out, defaults to &lsquo;main&rsquo; if no other field is defined.</p>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tag to check out, takes precedence over Branch.</p>
</td>
</tr>
<tr>
<td>
<code>commit</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Commit SHA to check out, takes precedence over all reference fields.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.LandscapeKitConfiguration">LandscapeKitConfiguration
</h3>
<p>
<p>LandscapeKitConfiguration contains configuration for the Gardener Landscape Kit.</p>
</p>
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
<code>ocm</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.OCMConfig">
OCMConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>OCM is the configuration for the OCM version processing.</p>
</td>
</tr>
<tr>
<td>
<code>git</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.GitRepository">
GitRepository
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Git is the configuration for the Git repository.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.OCMComponent">OCMComponent
</h3>
<p>
(<em>Appears on:</em>
<a href="#landscape.config.gardener.cloud/v1alpha1.OCMConfig">OCMConfig</a>)
</p>
<p>
<p>OCMComponent specifies a OCM component.</p>
</p>
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
<code>name</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.OCMConfig">OCMConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#landscape.config.gardener.cloud/v1alpha1.LandscapeKitConfiguration">LandscapeKitConfiguration</a>, 
<a href="#landscape.config.gardener.cloud/v1alpha1.OCMConfiguration">OCMConfiguration</a>)
</p>
<p>
<p>OCMConfig contains information about root component.</p>
</p>
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
<code>repositories</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Repositories is a map from repository name to URL.</p>
</td>
</tr>
<tr>
<td>
<code>rootComponent</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.OCMComponent">
OCMComponent
</a>
</em>
</td>
<td>
<p>RootComponent is the configuration of the root component.</p>
</td>
</tr>
<tr>
<td>
<code>originalRefs</code></br>
<em>
bool
</em>
</td>
<td>
<p>OriginalRefs is a flag to output original image references in the image vectors.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.OCMConfiguration">OCMConfiguration
</h3>
<p>
<p>OCMConfiguration contains information about root component.</p>
</p>
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
<code>OCMConfig</code></br>
<em>
<a href="#landscape.config.gardener.cloud/v1alpha1.OCMConfig">
OCMConfig
</a>
</em>
</td>
<td>
<p>
(Members of <code>OCMConfig</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="landscape.config.gardener.cloud/v1alpha1.PathConfiguration">PathConfiguration
</h3>
<p>
(<em>Appears on:</em>
<a href="#landscape.config.gardener.cloud/v1alpha1.GitRepository">GitRepository</a>)
</p>
<p>
<p>PathConfiguration contains path configuration within the Git repository.</p>
</p>
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
<code>base</code></br>
<em>
string
</em>
</td>
<td>
<p>Base is the relative path to the base directory within the Git repository.</p>
</td>
</tr>
<tr>
<td>
<code>landscape</code></br>
<em>
string
</em>
</td>
<td>
<p>Landscape is the relative path to the landscape directory within the Git repository.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
