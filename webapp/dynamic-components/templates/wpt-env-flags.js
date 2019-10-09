/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
WPTEnvironmentFlags is a class containing default enviroment wpt.fyi
feature flags.
*/
const WPTEnvironmentFlags = class WPTEnvironmentFlags {}
{{range .Flags}}
Object.defineProperty(
  WPTEnvironmentFlags,
  '{{.Name}}',
  {
    writable: false,
    configurable: false,
    value: {{.Enabled}},
  }
);
{{end}}

export { WPTEnvironmentFlags };