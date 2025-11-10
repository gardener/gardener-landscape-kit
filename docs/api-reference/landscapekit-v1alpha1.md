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
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
