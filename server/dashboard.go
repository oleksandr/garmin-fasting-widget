package main

import (
	"html/template"
	"net/http"
)

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Garmin Fasting Dashboard | fasting.home.lobunets.me</title>
    <style>
        :root {
            --bg-base: #0a0a0c;
            --bg-card: #141418;
            --bg-card-hover: #1c1c22;
            --border: #23232a;
            --cyan: #00d2ff;
            --cyan-glow: rgba(0, 210, 255, 0.15);
            --emerald: #30d158;
            --emerald-glow: rgba(48, 209, 88, 0.15);
            --amber: #ff9f0a;
            --amber-glow: rgba(255, 159, 10, 0.15);
            --text-main: #f5f5f7;
            --text-muted: #8e8e93;
            --text-sub: #636366;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg-base);
            color: var(--text-main);
            font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            min-height: 100vh;
            padding: 32px 20px;
            display: flex;
            flex-direction: column;
            align-items: center;
        }

        .container {
            width: 100%;
            max-width: 1080px;
        }

        /* Header */
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 28px;
            padding-bottom: 20px;
            border-bottom: 1px solid var(--border);
            flex-wrap: wrap;
            gap: 16px;
        }

        .header-title {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .header-logo {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: linear-gradient(135deg, var(--cyan), #007aff);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            box-shadow: 0 0 16px var(--cyan-glow);
        }

        h1 {
            font-size: 22px;
            font-weight: 700;
            letter-spacing: -0.5px;
        }

        .host-badge {
            font-size: 13px;
            color: var(--text-muted);
            margin-top: 2px;
            font-family: monospace;
        }

        .header-meta {
            display: flex;
            align-items: center;
            gap: 14px;
        }

        .status-pill {
            display: flex;
            align-items: center;
            gap: 6px;
            background: rgba(48, 209, 88, 0.1);
            color: var(--emerald);
            border: 1px solid rgba(48, 209, 88, 0.3);
            border-radius: 20px;
            padding: 4px 12px;
            font-size: 12px;
            font-weight: 600;
            letter-spacing: 0.5px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background-color: var(--emerald);
            box-shadow: 0 0 8px var(--emerald);
        }

        .server-time {
            color: var(--text-muted);
            font-size: 13px;
            font-family: monospace;
        }

        /* Active Fast Banner */
        .active-banner {
            background: linear-gradient(135deg, rgba(0, 210, 255, 0.1), rgba(0, 122, 255, 0.05));
            border: 1px solid rgba(0, 210, 255, 0.4);
            border-radius: 16px;
            padding: 20px 24px;
            margin-bottom: 28px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            box-shadow: 0 4px 24px var(--cyan-glow);
            flex-wrap: wrap;
            gap: 16px;
        }

        .active-left {
            display: flex;
            align-items: center;
            gap: 16px;
        }

        .active-pulse {
            width: 14px;
            height: 14px;
            border-radius: 50%;
            background: var(--cyan);
            box-shadow: 0 0 12px var(--cyan);
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(0, 210, 255, 0.7); }
            70% { transform: scale(1.05); box-shadow: 0 0 0 10px rgba(0, 210, 255, 0); }
            100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(0, 210, 255, 0); }
        }

        .active-title {
            font-size: 18px;
            font-weight: 700;
            color: var(--cyan);
        }

        .active-sub {
            color: var(--text-muted);
            font-size: 13px;
            margin-top: 2px;
        }

        .active-duration {
            font-size: 26px;
            font-weight: 700;
            font-family: monospace;
            color: #ffffff;
            letter-spacing: 0.5px;
        }

        /* KPI Stat Grid */
        .grid-stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
            gap: 16px;
            margin-bottom: 28px;
        }

        .stat-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 18px;
            transition: transform 0.15s, background-color 0.15s, border-color 0.15s;
        }

        .stat-card:hover {
            background-color: var(--bg-card-hover);
            border-color: #33333d;
            transform: translateY(-2px);
        }

        .stat-label {
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--text-muted);
            font-weight: 600;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .stat-value {
            font-size: 26px;
            font-weight: 700;
            color: #ffffff;
            letter-spacing: -0.5px;
            margin-bottom: 4px;
        }

        .stat-value.cyan { color: var(--cyan); }
        .stat-value.emerald { color: var(--emerald); }
        .stat-value.amber { color: var(--amber); }

        .stat-desc {
            font-size: 12px;
            color: var(--text-sub);
        }

        /* Main Content Layout */
        .content-grid {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 20px;
            margin-bottom: 28px;
        }

        @media (max-width: 840px) {
            .content-grid {
                grid-template-columns: 1fr;
            }
        }

        .panel {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 22px;
        }

        .panel-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 18px;
        }

        .panel-title {
            font-size: 16px;
            font-weight: 600;
            letter-spacing: -0.3px;
        }

        /* 7-Day Visual Overview */
        .week-chart {
            display: flex;
            justify-content: space-between;
            align-items: flex-end;
            height: 140px;
            padding: 10px 4px 0 4px;
            border-bottom: 1px dashed var(--border);
            position: relative;
        }

        .target-guide {
            position: absolute;
            left: 0;
            right: 0;
            bottom: calc((16 / 24) * 100%);
            border-top: 1px dashed rgba(0, 210, 255, 0.35);
            pointer-events: none;
        }

        .target-guide::after {
            content: "16h Goal";
            position: absolute;
            right: 2px;
            top: -16px;
            font-size: 10px;
            color: var(--cyan);
            font-weight: 600;
        }

        .bar-col {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
            flex: 1;
        }

        .bar-wrap {
            height: 110px;
            display: flex;
            align-items: flex-end;
            width: 100%;
            justify-content: center;
        }

        .bar {
            width: 24px;
            border-radius: 6px 6px 2px 2px;
            background: var(--cyan);
            transition: height 0.3s;
            position: relative;
        }

        .bar.goal-achieved {
            background: linear-gradient(180deg, var(--emerald), #228b22);
            box-shadow: 0 0 10px var(--emerald-glow);
        }

        .bar.zero-day {
            background: rgba(255, 255, 255, 0.05);
            height: 4px !important;
        }

        .bar-label {
            font-size: 11px;
            color: var(--text-muted);
            font-weight: 500;
        }

        .bar-hours {
            font-size: 10px;
            color: var(--text-sub);
            font-family: monospace;
        }

        /* Table */
        .table-wrap {
            width: 100%;
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 13px;
            text-align: left;
        }

        th {
            color: var(--text-muted);
            font-weight: 600;
            padding: 10px 12px;
            border-bottom: 1px solid var(--border);
            text-transform: uppercase;
            font-size: 11px;
            letter-spacing: 0.5px;
        }

        td {
            padding: 12px;
            border-bottom: 1px solid #1a1a20;
            color: #d1d1d6;
        }

        tr:last-child td {
            border-bottom: none;
        }

        tr:hover td {
            background-color: var(--bg-card-hover);
        }

        .badge {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 600;
        }

        .badge-goal {
            background: rgba(48, 209, 88, 0.15);
            color: var(--emerald);
            border: 1px solid rgba(48, 209, 88, 0.3);
        }

        .badge-completed {
            background: rgba(0, 210, 255, 0.15);
            color: var(--cyan);
            border: 1px solid rgba(0, 210, 255, 0.3);
        }

        .badge-active {
            background: rgba(255, 159, 10, 0.15);
            color: var(--amber);
            border: 1px solid rgba(255, 159, 10, 0.3);
        }

        .code-box {
            background: #0d0d10;
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 10px 12px;
            font-family: monospace;
            font-size: 12px;
            color: #a0a0ab;
            word-break: break-all;
            margin-top: 10px;
        }

        .empty-state {
            padding: 32px 16px;
            text-align: center;
            color: var(--text-muted);
            font-size: 14px;
        }

        footer {
            margin-top: 32px;
            padding-top: 20px;
            border-top: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
            color: var(--text-sub);
            font-size: 12px;
            flex-wrap: wrap;
            gap: 12px;
        }

        footer a {
            color: var(--cyan);
            text-decoration: none;
        }

        footer a:hover {
            text-decoration: underline;
        }
    </style>
    <script>
        // Auto refresh dashboard every 60 seconds
        setTimeout(() => {
            window.location.reload();
        }, 60000);
    </script>
</head>
<body>
    <div class="container">
        <!-- Top Header -->
        <header>
            <div class="header-title">
                <div class="header-logo">⏳</div>
                <div>
                    <h1>Garmin Fasting Tracker</h1>
                    <div class="host-badge">fasting.home.lobunets.me</div>
                </div>
            </div>
            <div class="header-meta">
                <div class="status-pill">
                    <div class="status-dot"></div>
                    ONLINE
                </div>
                <div class="server-time">{{.ServerTime}}</div>
            </div>
        </header>

        <!-- Active Fast Callout (if active) -->
        {{if .ActiveFast}}
        <div class="active-banner">
            <div class="active-left">
                <div class="active-pulse"></div>
                <div>
                    <div class="active-title">Fast in Progress</div>
                    <div class="active-sub">Started at {{.ActiveFast.StartTime}} (Target: {{.ActiveFast.TargetStr}}) • {{.ActiveFast.Device}}</div>
                </div>
            </div>
            <div class="active-duration">{{.ActiveFast.DurationStr}}</div>
        </div>
        {{end}}

        <!-- Primary KPI Metrics -->
        <div class="grid-stats">
            <div class="stat-card">
                <div class="stat-label">Streak <span>🔥</span></div>
                <div class="stat-value emerald">{{.Streak}} Days</div>
                <div class="stat-desc">Consecutive goals met</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Total Fasts <span>⏱️</span></div>
                <div class="stat-value cyan">{{.TotalFasts}}</div>
                <div class="stat-desc">Completed sessions</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Total Hours <span>⏳</span></div>
                <div class="stat-value">{{.TotalHours}}h</div>
                <div class="stat-desc">Time spent fasting</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Average <span>📊</span></div>
                <div class="stat-value">{{.AverageHours}}h</div>
                <div class="stat-desc">Mean fast duration</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Longest <span>🏆</span></div>
                <div class="stat-value amber">{{.LongestHours}}h</div>
                <div class="stat-desc">Personal best</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Success Rate <span>🎯</span></div>
                <div class="stat-value emerald">{{.SuccessRate}}</div>
                <div class="stat-desc">Achieved target goal</div>
            </div>
        </div>

        <!-- Content Grid: Weekly Activity + API Integration -->
        <div class="content-grid">
            <!-- 7-Day Activity Chart -->
            <div class="panel">
                <div class="panel-header">
                    <div class="panel-title">Weekly Activity (Last 7 Days)</div>
                </div>
                <div class="week-chart">
                    <div class="target-guide"></div>
                    {{range .WeeklyOverview}}
                    <div class="bar-col">
                        <div class="bar-wrap">
                            {{if eq .Hours 0.0}}
                            <div class="bar zero-day" style="height: 4px;" title="{{.DayName}}: 0h"></div>
                            {{else}}
                            <div class="bar {{if .Achieved}}goal-achieved{{end}}" style="height: {{.HeightPct}}%;" title="{{.DayName}}: {{printf "%.1f" .Hours}}h"></div>
                            {{end}}
                        </div>
                        <div class="bar-label">{{.DayName}}</div>
                        <div class="bar-hours">{{if eq .Hours 0.0}}-{{else}}{{printf "%.0f" .Hours}}h{{end}}</div>
                    </div>
                    {{end}}
                </div>
            </div>

            <!-- Mobile App & API Quick Reference -->
            <div class="panel">
                <div class="panel-header">
                    <div class="panel-title">API Integration</div>
                    <span class="badge badge-completed">Endpoint: /api</span>
                </div>
                <p style="font-size: 13px; color: var(--text-muted); line-height: 1.5;">
                    Post sync payloads from Garmin Connect Mobile or smartphone shortcuts:
                </p>
                <div class="code-box">
                    POST https://fasting.home.lobunets.me/api<br>
                    Content-Type: application/json<br>
                    Authorization: Bearer &lt;TOKEN&gt;
                </div>
                <p style="font-size: 12px; color: var(--text-sub); margin-top: 10px;">
                    JSON endpoint: <a href="/api" style="color: var(--cyan); text-decoration: none;">GET /api</a>
                </p>
            </div>
        </div>

        <!-- History Table -->
        <div class="panel">
            <div class="panel-header">
                <div class="panel-title">Recent Fasting History</div>
                <span style="font-size: 12px; color: var(--text-muted);">Last 30 Sessions</span>
            </div>

            {{if .RecentFasts}}
            <div class="table-wrap">
                <table>
                    <thead>
                        <tr>
                            <th>Start</th>
                            <th>End</th>
                            <th>Duration</th>
                            <th>Target</th>
                            <th>Status</th>
                            <th>Device</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .RecentFasts}}
                        <tr>
                            <td>
                                <strong style="color: #fff;">{{.StartDate}}</strong><br>
                                <span style="font-size: 11px; color: var(--text-muted);">{{.StartTime}}</span>
                            </td>
                            <td>
                                {{if eq .EndTime "--:--"}}
                                <span style="color: var(--amber);">--:--</span>
                                {{else}}
                                <span>{{.EndDate}}</span><br>
                                <span style="font-size: 11px; color: var(--text-muted);">{{.EndTime}}</span>
                                {{end}}
                            </td>
                            <td>
                                <strong style="font-family: monospace; font-size: 14px; color: #fff;">{{.DurationStr}}</strong>
                            </td>
                            <td>{{.TargetStr}}</td>
                            <td>
                                <span class="badge {{.StatusBadgeClass}}">{{.StatusBadgeText}}</span>
                            </td>
                            <td style="color: var(--text-muted); font-size: 12px;">{{.Device}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{else}}
            <div class="empty-state">
                <p>No fasting records synced yet.</p>
                <p style="margin-top: 6px; font-size: 12px; color: var(--text-sub);">
                    Sync your Garmin Instinct 3 AMOLED or post records to <code>/api</code> to get started.
                </p>
            </div>
            {{end}}
        </div>

        <!-- Footer -->
        <footer>
            <div>Garmin Fasting Companion &bull; Self-Hosted on ZimaBoard &bull; Pure Go &bull; SQLite</div>
            <div>Uptime: {{.ServerUptime}} &bull; <a href="/api">JSON API</a></div>
        </footer>
    </div>
</body>
</html>
`

var parsedDashboardTmpl = template.Must(template.New("dashboard").Parse(dashboardHTML))

func renderDashboard(w http.ResponseWriter, data DashboardData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = parsedDashboardTmpl.Execute(w, data)
}
