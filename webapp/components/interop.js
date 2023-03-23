/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { InteropDashboard } from './interop-dashboard.js';
import { InteropFeatureChart } from './interop-feature-chart.js';
import { InteropSummary } from './interop-summary.js';

window.customElements.define(InteropFeatureChart.is, InteropFeatureChart);
window.customElements.define(InteropSummary.is, InteropSummary);
window.customElements.define(InteropDashboard.is, InteropDashboard);
