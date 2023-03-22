import { InteropDashboard } from "./interop-dashboard.js";
import { InteropFeatureChart } from "./interop-feature-chart.js";
import { InteropSummary } from "./interop-summary.js";

window.customElements.define(InteropFeatureChart.is, InteropFeatureChart);
window.customElements.define(InteropSummary.is, InteropSummary);
window.customElements.define(InteropDashboard.is, InteropDashboard);