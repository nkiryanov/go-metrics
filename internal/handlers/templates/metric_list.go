package templates

import (
	"html/template"
)

var MetricList = template.Must(template.New("listTemplate").Parse(`<!DOCTYPE html>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics List</title>
    <style>
        body {
            font-family: Arial, sans-serif; /* Primary font and fallback */
        }
        h1 {
            font-size: 24px;
            color: #333;
            text-align: center;
            margin: 20px 0;
        }
        table {
            width: 50%;
            margin: 20px auto;
            border-collapse: collapse;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
    </style>
    </style>
</head>
<body>
    <h1>Metrics List</h1>
    {{- if . -}}
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Value</th>
                <th>Type</th>
            </tr>
        </thead>
        <tbody>
            {{- range . -}}
            <tr>
                <td>{{ .MName }}</td>
                <td>{{ .MValue }}</td>
                <td>{{ .MType }}</td>
            </tr>
            {{- end -}}
        </tbody>
    </table>
    {{- else -}}
    <p>No metrics found</p>
    {{- end -}}
</body>
</html>
`))
