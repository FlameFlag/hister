{{ define "main" }}
<div class="section">
    <h1>API documentation</h1>
    <ul>
        {{ range .Endpoints }}
        <li><a href="#{{ Replace .Name " " "_" | ToLower}}_{{ .Method }}">{{ .Name }}</a></li>
        {{ end }}
    </ul>
    {{ range .Endpoints }}
    <div class="container" id="{{ Replace .Name " " "_" | ToLower}}_{{ .Method }}">
        <div>
            <h3>{{ .Name }}</h3>
            <h2 class="success"><code>{{ .Method }}</code><code>{{ .Path }}</code>{{ if .CSRFRequired }}<span class="small grey"> CSRF</span>{{ end }}</h2>
            <p>{{ .Description }}</p>
            <hr />
            {{ if .Args }}
            <h4>Arguments</h4>
            <table>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Required</th>
                    <th>Description</th>
                </tr>
                {{ range .Args }}
                <tr>
                    <td><code>{{ .Name }}</code></td>
                    <td><code>{{ .Type }}</code></td>
                    <td>{{ .Required }}</td>
                    <td>{{ .Description }}</td>
                {{ end }}
            </table>
            {{ else }}
            <h5>No arguments available for this endpoint</h5>
            {{ end }}
        </div>
    </div>
    {{ end }}
</div>
{{ end }}
