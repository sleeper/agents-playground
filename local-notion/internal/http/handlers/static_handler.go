package handlers

import "net/http"

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Local Notion</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
            margin: 0;
            padding: 3rem 1.5rem;
            background: #0f172a;
            color: #e2e8f0;
            display: flex;
            justify-content: center;
        }
        main {
            max-width: 720px;
            background: rgba(15, 23, 42, 0.85);
            border: 1px solid rgba(148, 163, 184, 0.25);
            border-radius: 18px;
            padding: 2.5rem;
            box-shadow: 0 30px 60px -35px rgba(15, 23, 42, 0.9);
        }
        h1 {
            margin-top: 0;
            font-size: 2.5rem;
            letter-spacing: -0.03em;
        }
        p {
            line-height: 1.6;
        }
        code {
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
            background: rgba(30, 41, 59, 0.7);
            padding: 0.25rem 0.5rem;
            border-radius: 6px;
        }
        a {
            color: #38bdf8;
        }
    </style>
</head>
<body>
    <main>
        <h1>Local Notion</h1>
        <p>
            Your self-hosted workspace is running. The production web client is not yet bundled,
            but you can start the React development server from <code>web/</code> while the Go
            backend stays online.
        </p>
        <p>
            From another terminal run:
        </p>
        <pre><code>cd web
npm install
npm run dev</code></pre>
        <p>
            The API remains available under <code>/api</code>. See the README for detailed usage
            instructions.
        </p>
    </main>
</body>
</html>`

// IndexHandler responds with a minimal HTML shell so navigating to the root
// path succeeds even when the frontend bundle is not built yet.
func IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(indexHTML))
	}
}

// FaviconHandler returns a 204 response so browsers do not log a noisy 404 when
// requesting /favicon.ico in development.
func FaviconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}
