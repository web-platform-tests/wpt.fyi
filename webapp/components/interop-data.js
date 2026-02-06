// This file should match the data in webapp/static/interop-data.json.
// The JSON file is used by some Mozilla infrastructure, so the file should
// not be deleted and should match the data in this file.
export const interopData = {
  'valid_years': ['2021', '2022', '2023', '2024', '2025'],
  'valid_mobile_years': ['2024', '2025'],
  '2021': {
    'table_sections': [
      {
        'name': '2021 Focus Areas',
        'rows': [
          'interop-2021-aspect-ratio',
          'interop-2021-flexbox',
          'interop-2021-grid',
          'interop-2021-transforms',
          'interop-2021-position-sticky'
        ],
        'score_as_group': false
      }
    ],
    /**
     * Interop scores are "frozen" after the end of the year.
     * Once an interop year is completed, results are generated one more time
     * from the results-analysis script for the full year, and those scores
     * are placed for reference in the webapp/static directory. The score
     * is no longer updated and is referenced from this location.
     * More information at https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': '/static/interop-2021-{stable|experimental}.csv',
    'summary_feature_name': 'summary',
    'matrix_url': 'https://matrix.to/#/#interop20xx:matrix.org?web-instance%5Belement.io%5D=app.element.io',
    'focus_areas': {
      'interop-2021-aspect-ratio': {
        'description': 'Aspect Ratio',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
        'spec': 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'labels': [
          'interop-2021-aspect-ratio'
        ]
      },
      'interop-2021-flexbox': {
        'description': 'Flexbox',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox',
        'labels': [
          'interop-2021-flexbox'
        ]
      },
      'interop-2021-grid': {
        'description': 'Grid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/grid',
        'spec': 'https://drafts.csswg.org/css-grid-1/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid',
        'labels': [
          'interop-2021-grid'
        ]
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'labels': [
          'interop-2021-position-sticky'
        ]
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'labels': [
          'interop-2021-transforms'
        ]
      }
    }
  },
  '2022': {
    'table_sections': [
      {
        'name': '2022 Focus Areas',
        'rows': [
          'interop-2021-aspect-ratio',
          'interop-2022-cascade',
          'interop-2022-color',
          'interop-2022-contain',
          'interop-2022-dialog',
          'interop-2021-flexbox',
          'interop-2022-forms',
          'interop-2021-grid',
          'interop-2022-scrolling',
          'interop-2021-position-sticky',
          'interop-2022-subgrid',
          'interop-2022-text',
          'interop-2021-transforms',
          'interop-2022-viewport',
          'interop-2022-webcompat'
        ],
        'score_as_group': false
      },
      {
        'name': '2022 Investigations',
        'rows': [
          'Editing, contenteditable, and execCommand',
          'Pointer and Mouse Events',
          'Viewport Measurement'
        ],
        'score_as_group': true
      }
    ],
    'investigation_scores': [
      {
        'name': 'Editing, contenteditable, and execCommand',
        'url': 'https://github.com/web-platform-tests/interop-2022-editing',
        'scores_over_time': [
          { 'date': '2022-10-22', 'score': 360 },
          { 'date': '2022-11-25', 'score': 460 },
          { 'date': '2022-12-15', 'score': 520 }
        ]
      },
      {
        'name': 'Pointer and Mouse Events',
        'url': 'https://github.com/web-platform-tests/interop-2022-pointer',
        'scores_over_time': [
          { 'date': '2022-12-01', 'score': 790 },
          { 'date': '2022-12-14', 'score': 1000 }
        ]
      },
      {
        'name': 'Viewport Measurement',
        'url': 'https://github.com/web-platform-tests/interop-2022-viewport',
        'scores_over_time': [
          { 'date': '2022-09-28', 'score': 600 },
          { 'date': '2022-12-14', 'score': 900 }
        ]
      }
    ],
    'investigation_weight': 0.1,
    /**
     * Interop scores are "frozen" after the end of the year.
     * Once an interop year is completed, results are generated one more time
     * from the results-analysis script for the full year, and those scores
     * are placed for reference in the webapp/static directory. The score
     * is no longer updated and is referenced from this location.
     * More information at https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': '/static/interop-2022-{stable|experimental}.csv',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas': {
      'interop-2021-aspect-ratio': {
        'description': 'Aspect Ratio',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
        'spec': 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'labels': [
          'interop-2021-aspect-ratio'
        ]
      },
      'interop-2021-flexbox': {
        'description': 'Flexbox',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox',
        'labels': [
          'interop-2021-flexbox'
        ]
      },
      'interop-2021-grid': {
        'description': 'Grid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/grid',
        'spec': 'https://drafts.csswg.org/css-grid-1/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid',
        'labels': [
          'interop-2021-grid'
        ]
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'labels': [
          'interop-2021-position-sticky'
        ]
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'labels': [
          'interop-2021-transforms'
        ]
      },
      'interop-2022-cascade': {
        'description': 'Cascade Layers',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@layer',
        'spec': 'https://drafts.csswg.org/css-cascade/#layering',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade',
        'labels': [
          'interop-2022-cascade'
        ]
      },
      'interop-2022-color': {
        'description': 'Color Spaces and Functions',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/color_value',
        'spec': 'https://drafts.csswg.org/css-color/',
        'tests': '/results/css/css-color?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color',
        'labels': [
          'interop-2022-color'
        ]
      },
      'interop-2022-contain': {
        'countsTowardScore': true,
        'description': 'Containment',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/contain',
        'spec': 'https://drafts.csswg.org/css-contain/#contain-property',
        'tests': '/results/css/css-contain?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain',
        'labels': [
          'interop-2022-contain'
        ]
      },
      'interop-2022-dialog': {
        'description': 'Dialog Element',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog',
        'labels': [
          'interop-2022-dialog'
        ]
      },
      'interop-2022-forms': {
        'description': 'Forms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
        'spec': 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms',
        'labels': [
          'interop-2022-forms'
        ]
      },
      'interop-2022-scrolling': {
        'description': 'Scrolling',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/overflow',
        'spec': 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
        'labels': [
          'interop-2022-scrolling'
        ]
      },
      'interop-2022-subgrid': {
        'description': 'Subgrid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
        'spec': 'https://drafts.csswg.org/css-grid-2/#subgrids',
        'tests': '/results/css/css-grid/subgrid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-subgrid',
        'labels': [
          'interop-2022-subgrid'
        ]
      },
      'interop-2022-text': {
        'description': 'Typography and Encodings',
        'countsTowardScore': true,
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text',
        'labels': [
          'interop-2022-text'
        ]
      },
      'interop-2022-viewport': {
        'description': 'Viewport Units',
        'countsTowardScore': true,
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-values/#viewport-relative-units',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport',
        'labels': [
          'interop-2022-viewport'
        ]
      },
      'interop-2022-webcompat': {
        'description': 'Web Compat',
        'countsTowardScore': true,
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
        'labels': [
          'interop-2022-webcompat'
        ]
      }
    }
  },
  '2023': {
    'table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2023-cssborderimage',
          'interop-2023-color',
          'interop-2023-container',
          'interop-2023-contain',
          'interop-2023-mathfunctions',
          'interop-2023-pseudos',
          'interop-2023-property',
          'interop-2023-flexbox',
          'interop-2023-fonts',
          'interop-2023-forms',
          'interop-2023-grid',
          'interop-2023-has',
          'interop-2023-inert',
          'interop-2023-cssmasking',
          'interop-2023-mediaqueries',
          'interop-2023-modules',
          'interop-2023-motion',
          'interop-2023-offscreencanvas',
          'interop-2023-events',
          'interop-2022-scrolling',
          'interop-2022-subgrid',
          'interop-2021-transforms',
          'interop-2023-url',
          'interop-2023-webcodecs',
          'interop-2023-webcompat',
          'interop-2023-webcomponents'
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility Testing',
          'Mobile Testing'
        ],
        'score_as_group': true
      },
      {
        'name': 'Previous Focus Areas',
        'rows': [
          'interop-2021-aspect-ratio',
          'interop-2022-cascade',
          'interop-2022-dialog',
          'interop-2021-position-sticky',
          'interop-2022-text',
          'interop-2022-viewport',
          'interop-2022-webcompat'
        ],
        'score_as_group': false
      },
      {
        'name': 'Previous Investigations',
        'rows': [
          'Editing, contenteditable, and execCommand',
          'Pointer and Mouse Events',
          'Viewport Measurement'
        ],
        'previous_investigation': true,
        'score_as_group': true
      }
    ],
    'investigation_scores': [
      {
        'name': 'Accessibility Testing',
        'url': 'https://github.com/web-platform-tests/interop-accessibility',
        'scores_over_time': [
          { 'date': '2023-03-08', 'score': 600 },
          { 'date': '2023-05-13', 'score': 700 },
          { 'date': '2023-06-27', 'score': 780 },
          { 'date': '2023-09-05', 'score': 860 },
          { 'date': '2023-09-27', 'score': 870 },
          { 'date': '2023-10-12', 'score': 910 },
          { 'date': '2023-10-13', 'score': 920 },
          { 'date': '2023-11-03', 'score': 950 },
          { 'date': '2023-11-14', 'score': 980 },
          { 'date': '2023-11-19', 'score': 1000 }
        ]
      },
      {
        'name': 'Mobile Testing',
        'url': 'https://github.com/web-platform-tests/interop-2023-mobile-testing',
        'scores_over_time': [
          { 'date': '2023-06-20', 'score': 400 },
          { 'date': '2023-09-26', 'score': 600 },
          { 'date': '2023-10-24', 'score': 700 }
        ]
      }
    ],
    'investigation_weight': 0.0,
    /**
     * More information on results generation at
     * https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': '/static/interop-2023-{stable|experimental}.csv',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas': {
      'interop-2021-aspect-ratio': {
        'description': 'Aspect Ratio',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
        'spec': 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'countsTowardScore': false,
        'labels': [
          'interop-2021-aspect-ratio'
        ]
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'countsTowardScore': false,
        'labels': [
          'interop-2021-position-sticky'
        ]
      },
      'interop-2022-cascade': {
        'description': 'Cascade Layers',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@layer',
        'spec': 'https://drafts.csswg.org/css-cascade/#layering',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-cascade'
        ]
      },
      'interop-2022-dialog': {
        'description': 'Dialog Element',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-dialog'
        ]
      },
      'interop-2022-text': {
        'description': 'Typography and Encodings',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/length#relative_length_units_based_on_viewport',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-text'
        ]
      },
      'interop-2022-viewport': {
        'description': 'Viewport Units',
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-values/#viewport-relative-units',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-viewport'
        ]
      },
      'interop-2022-webcompat': {
        'description': 'Web Compat 2022',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-webcompat'
        ]
      },
      'interop-2023-cssborderimage': {
        'description': 'Border Image',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/border-image',
        'spec': 'https://www.w3.org/TR/css-backgrounds-3/#the-border-image',
        'tests': '/results/css/css-backgrounds?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssborderimage',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-cssborderimage'
        ]
      },
      'interop-2023-color': {
        'description': 'Color Spaces and Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/color_value',
        'spec': 'https://w3c.github.io/csswg-drafts/css-color/#color-syntax',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color%20or%20label%3Ainterop-2023-color',
        'countsTowardScore': true,
        'labels': [
          'interop-2022-color',
          'interop-2023-color'
        ]
      },
      'interop-2023-container': {
        'description': 'Container Queries',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Container_Queries',
        'spec': 'https://drafts.csswg.org/css-contain-3/#container-queries',
        'tests': '/results/css/css-contain/container-queries?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-container',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-container'
        ]
      },
      'interop-2023-contain': {
        'description': 'Containment',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/contain',
        'spec': 'https://drafts.csswg.org/css-contain/#contain-property',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain%20or%20label%3Ainterop-2023-contain',
        'countsTowardScore': true,
        'labels': [
          'interop-2022-contain',
          'interop-2023-contain'
        ]
      },
      'interop-2023-pseudos': {
        'description': 'CSS Pseudo-classes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Pseudo-classes',
        'spec': 'https://drafts.csswg.org/selectors/',
        'tests': '/results/css/selectors?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-pseudos',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-pseudos'
        ]
      },
      'interop-2023-property': {
        'description': 'Custom Properties',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@property',
        'spec': 'https://drafts.css-houdini.org/css-properties-values-api/',
        'tests': '/results/css/css-properties-values-api?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-property',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-property'
        ]
      },
      'interop-2023-flexbox': {
        'description': 'Flexbox',
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox',
        'countsTowardScore': true,
        'labels': [
          'interop-2021-flexbox',
          'interop-2023-flexbox'
        ]
      },
      'interop-2023-fonts': {
        'description': 'Font Feature Detection and Palettes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/font-palette',
        'spec': 'https://drafts.csswg.org/css-fonts-4/#font-palette-prop',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-fonts',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-fonts'
        ]
      },
      'interop-2023-forms': {
        'description': 'Forms',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
        'spec': 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms%20or%20label%3Ainterop-2023-forms',
        'countsTowardScore': true,
        'labels': [
          'interop-2022-forms',
          'interop-2023-forms'
        ]
      },
      'interop-2023-grid': {
        'description': 'Grid',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout',
        'spec': 'https://drafts.csswg.org/css-grid/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid',
        'countsTowardScore': true,
        'labels': [
          'interop-2021-grid',
          'interop-2023-grid'
        ]
      },
      'interop-2023-has': {
        'description': ':has()',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/:has',
        'spec': 'https://drafts.csswg.org/selectors-4/#relational',
        'tests': '/results/css/selectors?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-has',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-has'
        ]
      },
      'interop-2023-inert': {
        'description': 'Inert',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Global_attributes/inert',
        'spec': 'https://html.spec.whatwg.org/multipage/interaction.html#the-inert-attribute',
        'tests': '/results/inert?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-inert',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-inert'
        ]
      },
      'interop-2023-cssmasking': {
        'description': 'Masking',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Masking',
        'spec': 'https://drafts.fxtf.org/css-masking/',
        'tests': '/results/css/css-masking?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssmasking',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-cssmasking'
        ]
      },
      'interop-2023-mathfunctions': {
        'description': 'CSS Math Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Functions#math_functions',
        'spec': 'https://drafts.csswg.org/css-values-4/#math',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mathfunctions',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-mathfunctions'
        ]
      },
      'interop-2023-mediaqueries': {
        'description': 'Media Queries 4',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Media_Queries/Using_media_queries',
        'spec': 'https://www.w3.org/TR/mediaqueries-4/',
        'tests': '/results/css/mediaqueries?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mediaqueries',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-mediaqueries'
        ]
      },
      'interop-2023-modules': {
        'description': 'Modules',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules',
        'spec': 'https://tc39.es/proposal-import-assertions/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-modules',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-modules'
        ]
      },
      'interop-2023-motion': {
        'description': 'Motion Path',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Motion_Path',
        'spec': 'https://drafts.fxtf.org/motion-1/',
        'tests': '/results/css/motion?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-motion',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-motion'
        ]
      },
      'interop-2023-offscreencanvas': {
        'description': 'Offscreen Canvas',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/OffscreenCanvas',
        'spec': 'https://html.spec.whatwg.org/multipage/canvas.html#the-offscreencanvas-interface',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-offscreencanvas',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-offscreencanvas'
        ]
      },
      'interop-2023-events': {
        'description': 'Pointer and Mouse Events',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/Pointer_events',
        'spec': 'https://w3c.github.io/pointerevents/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-events',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-events'
        ]
      },
      'interop-2022-scrolling': {
        'description': 'Scrolling',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/overflow',
        'spec': 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
        'countsTowardScore': true,
        'labels': [
          'interop-2022-scrolling'
        ]
      },
      'interop-2022-subgrid': {
        'description': 'Subgrid',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
        'spec': 'https://drafts.csswg.org/css-grid-2/#subgrids',
        'tests': '/results/css/css-grid/subgrid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-subgrid',
        'countsTowardScore': true,
        'labels': [
          'interop-2022-subgrid'
        ]
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'countsTowardScore': true,
        'labels': [
          'interop-2021-transforms'
        ]
      },
      'interop-2023-url': {
        'description': 'URL',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/URL',
        'spec': 'https://url.spec.whatwg.org',
        'tests': '/results/url?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-url',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-url'
        ]
      },
      'interop-2023-webcompat': {
        'description': 'Web Compat 2023',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcompat',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-webcompat'
        ]
      },
      'interop-2023-webcodecs': {
        'description': 'Web Codecs (video)',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/WebCodecs_API',
        'spec': 'https://www.w3.org/TR/webcodecs/',
        'tests': '/results/webcodecs?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcodecs',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-webcodecs'
        ]
      },
      'interop-2023-webcomponents': {
        'description': 'Web Components',
        'mdn': 'https://developer.mozilla.org/docs/Web/Web_Components',
        'spec': 'https://www.w3.org/wiki/WebComponents/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcomponents',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-webcomponents'
        ]
      }
    }
  },
  '2024': {
    'browsers': ['chrome_canary', 'edge', 'firefox', 'safari'],
    'mobile_browsers': ['chrome_android', 'firefox_android'],
    'table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2024-accessibility',
          'interop-2024-nesting',
          'interop-2023-property',
          'interop-2024-dsd',
          'interop-2024-font-size-adjust',
          'interop-2024-websockets',
          'interop-2024-indexeddb',
          'interop-2024-layout',
          'interop-2023-events',
          'interop-2024-popover',
          'interop-2024-relative-color',
          'interop-2024-video-rvfc',
          'interop-2024-scrollbar',
          'interop-2024-starting-style-transition-behavior',
          'interop-2024-dir',
          'interop-2024-text-wrap',
          'interop-2023-url'
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility Testing',
          'Mobile Testing',
          'WebAssembly Testing'
        ],
        'score_as_group': true
      },
      {
        'name': 'Previous Focus Areas',
        'rows': [
          'interop-2021-aspect-ratio',
          'interop-2023-cssborderimage',
          'interop-2022-cascade',
          'interop-2023-color',
          'interop-2023-container',
          'interop-2023-contain',
          'interop-2023-mathfunctions',
          'interop-2023-pseudos',
          'interop-2022-dialog',
          'interop-2023-fonts',
          'interop-2023-forms',
          'interop-2023-has',
          'interop-2023-inert',
          'interop-2023-cssmasking',
          'interop-2023-mediaqueries',
          'interop-2023-modules',
          'interop-2023-motion',
          'interop-2023-offscreencanvas',
          'interop-2022-scrolling',
          'interop-2021-position-sticky',
          'interop-2021-transforms',
          'interop-2022-text',
          'interop-2022-viewport',
          'interop-2023-webcodecs',
          'interop-2022-webcompat',
          'interop-2023-webcompat',
          'interop-2023-webcomponents'
        ],
        'score_as_group': false
      }
    ],
    'investigation_scores': [
      {
        'name': 'Accessibility Testing',
        'url': 'https://github.com/web-platform-tests/interop-accessibility',
        'scores_over_time': [
          { 'date': '2024-04-02', 'score': 18 },
          { 'date': '2024-04-25', 'score': 33 },
          { 'date': '2024-06-28', 'score': 120 },
          { 'date': '2024-08-13', 'score': 242 },
          { 'date': '2024-10-01', 'score': 458 },
          { 'date': '2024-11-11', 'score': 558 },
        ]
      },
      {
        'name': 'Mobile Testing',
        'url': 'https://github.com/web-platform-tests/interop-mobile-testing',
        'scores_over_time': [
          { 'date': '2024-04-23', 'score': 130 },
          { 'date': '2024-10-22', 'score': 625 },
        ]
      },
      {
        'name': 'WebAssembly Testing',
        'url': 'https://github.com/web-platform-tests/interop-2024-wasm',
        'scores_over_time': [
          { 'date': '2024-11-19', 'score': 333 },
          { 'date': '2024-12-20', 'score': 666 },
        ]
      }
    ],
    'investigation_weight': 0.0,
    'mobile_table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2024-accessibility',
          'interop-2024-nesting',
          'interop-2023-property',
          'interop-2024-dsd',
          'interop-2024-font-size-adjust',
          'interop-2024-websockets',
          'interop-2024-indexeddb',
          'interop-2024-layout',
          'interop-2023-events',
          'interop-2024-popover',
          'interop-2024-relative-color',
          'interop-2024-video-rvfc',
          'interop-2024-scrollbar',
          'interop-2024-starting-style-transition-behavior',
          'interop-2024-dir',
          'interop-2024-text-wrap',
          'interop-2023-url',
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility Testing',
          'Mobile Testing',
          'WebAssembly Testing'
        ],
        'score_as_group': true
      }
    ],
    'mobile_focus_areas': [
      'interop-2024-accessibility',
      'interop-2024-nesting',
      'interop-2023-property',
      'interop-2024-dsd',
      'interop-2024-font-size-adjust',
      'interop-2024-websockets',
      'interop-2024-indexeddb',
      'interop-2024-layout',
      'interop-2023-events',
      'interop-2024-popover',
      'interop-2024-relative-color',
      'interop-2024-video-rvfc',
      'interop-2024-scrollbar',
      'interop-2024-starting-style-transition-behavior',
      'interop-2024-dir',
      'interop-2024-text-wrap',
      'interop-2023-url',
    ],
    /**
     * More information on results generation at
     * https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': '/static/interop-2024-{stable|experimental}.csv',
    'mobile_csv_url': '/static/interop-2024-mobile-experimental.csv',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas_description': 'https://github.com/web-platform-tests/interop/blob/main/2024/README.md',
    'focus_areas': {
      'interop-2021-aspect-ratio': {
        'description': 'Aspect Ratio',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
        'spec': 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'countsTowardScore': false,
        'labels': [
          'interop-2021-aspect-ratio'
        ]
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'mobile_tests': '/results/css/css-position/sticky?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'countsTowardScore': false,
        'labels': [
          'interop-2021-position-sticky'
        ]
      },
      'interop-2022-cascade': {
        'description': 'Cascade Layers',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@layer',
        'spec': 'https://drafts.csswg.org/css-cascade/#layering',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade',
        'mobile_tests': '/results/css/css-cascade?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-cascade',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-cascade'
        ]
      },
      'interop-2022-dialog': {
        'description': 'Dialog Element',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-dialog',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-dialog'
        ]
      },
      'interop-2022-text': {
        'description': 'Typography and Encodings',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/length#relative_length_units_based_on_viewport',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-text',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-text'
        ]
      },
      'interop-2022-viewport': {
        'description': 'Viewport Units',
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-values/#viewport-relative-units',
        'tests': '/results/css/css-values?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport',
        'mobile_tests': '/results/css/css-values?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-viewport',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-viewport'
        ]
      },
      'interop-2022-webcompat': {
        'description': 'Web Compat 2022',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-webcompat'
        ]
      },
      'interop-2024-accessibility': {
        'description': 'Accessibility',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Glossary/Accessible_name',
        'spec': '',
        'tests': '/results/?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-accessibility',
        'mobile_tests': '/results/?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-accessibility',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-accessibility'
        ]
      },
      'interop-2024-starting-style-transition-behavior': {
        'description': '@starting-style & transition-behavior',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/@starting-style',
        'spec': '',
        'tests': '/results/css?label=experimental&label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-starting-style%20or%20label%3Ainterop-2024-transition-behavior',
        'mobile_tests': '/results/css?label=experimental&label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-starting-style%20or%20label%3Ainterop-2024-transition-behavior',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-starting-style',
          'interop-2024-transition-behavior'
        ]
      },
      'interop-2024-dsd': {
        'description': 'Declarative Shadow DOM',
        'mdn': '',
        'spec': '',
        'tests': '/shadow-dom?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-dsd',
        'mobile_tests': '/shadow-dom?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-dsd',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-dsd'
        ]
      },
      'interop-2024-dir': {
        'description': 'Text Directionality',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/:dir',
        'spec': '',
        'tests': '/results/html/dom/elements/global-attributes?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-dir',
        'mobile_tests': '/results/html/dom/elements/global-attributes?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-dir',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-dir'
        ]
      },
      'interop-2024-font-size-adjust': {
        'description': 'font-size-adjust',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/font-size-adjust',
        'spec': '',
        'tests': '/results/css/css-fonts?label=experimental&label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-font-size-adjust',
        'mobile_tests': '/results/css/css-fonts?label=experimental&label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-font-size-adjust',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-font-size-adjust'
        ]
      },
      'interop-2024-websockets': {
        'description': 'HTTPS URLs for WebSocket',
        'mdn': '',
        'spec': 'https://websockets.spec.whatwg.org/ ',
        'tests': '/results/websockets?label=experimental&label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-websockets',
        'mobile_tests': '/results/websockets?label=experimental&label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-websockets',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-websockets'
        ]
      },
      'interop-2024-indexeddb': {
        'description': 'IndexedDB',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API/Using_IndexedDB',
        'spec': '',
        'tests': '/results/IndexedDB?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-indexeddb',
        'mobile_tests': '/results/IndexedDB?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-indexeddb',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-indexeddb'
        ]
      },
      'interop-2024-layout': {
        'description': 'Layout',
        'mdn': '',
        'spec': '',
        'tests': '/results/css?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox%20or%20label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid%20or%20label%3Ainterop-2022-subgrid',
        'mobile_tests': '/results/css?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox%20or%20label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid%20or%20label%3Ainterop-2022-subgrid',
        'countsTowardScore': true,
        'labels': [
          'interop-2021-flexbox',
          'interop-2021-grid',
          'interop-2022-subgrid',
          'interop-2023-flexbox',
          'interop-2023-grid'
        ]
      },
      'interop-2024-nesting': {
        'description': 'CSS Nesting',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_nesting',
        'spec': '',
        'tests': '/results/css?label=experimental&label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-nesting',
        'mobile_tests': '/results/css?label=experimental&label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-nesting',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-nesting'
        ]
      },
      'interop-2024-popover': {
        'description': 'Popover',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/Popover_API',
        'spec': '',
        'tests': '/results/html/semantics/popovers?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-popover',
        'mobile_tests': '/results/html/semantics/popovers?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-popover',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-popover'
        ]
      },
      'interop-2024-relative-color': {
        'description': 'Relative Color Syntax',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-color?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-relative-color',
        'mobile_tests': '/results/css/css-color?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-relative-color',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-relative-color'
        ]
      },
      'interop-2024-video-rvfc': {
        'description': 'requestVideoFrameCallback',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/HTMLVideoElement/requestVideoFrameCallback',
        'spec': '',
        'tests': '/results/video-rvfc?label=experimental&label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-video-rvfc',
        'mobile_tests': '/results/video-rvfc?label=experimental&label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-video-rvfc',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-video-rvfc'
        ]
      },
      'interop-2024-scrollbar': {
        'description': 'Scrollbar Styling',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/scrollbar-width',
        'spec': '',
        'tests': '/results/css?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-scrollbar',
        'mobile_tests': '/results/css?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-scrollbar',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-scrollbar'
        ]
      },
      'interop-2024-text-wrap': {
        'description': 'text-wrap: balance',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/text-wrap',
        'spec': '',
        'tests': '/results/css/css-text?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2024-text-wrap',
        'mobile_tests': '/results/css/css-text?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2024-text-wrap',
        'countsTowardScore': true,
        'labels': [
          'interop-2024-text-wrap'
        ]
      },
      'interop-2023-cssborderimage': {
        'description': 'Border Image',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/border-image',
        'spec': 'https://www.w3.org/TR/css-backgrounds-3/#the-border-image',
        'tests': '/results/css/css-backgrounds?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssborderimage',
        'mobile_tests': '/results/css/css-backgrounds?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-cssborderimage',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-cssborderimage'
        ]
      },
      'interop-2023-color': {
        'description': 'Color Spaces and Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/color_value',
        'spec': 'https://w3c.github.io/csswg-drafts/css-color/#color-syntax',
        'tests': '/results/css?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color%20or%20label%3Ainterop-2023-color',
        'mobile_tests': '/results/css?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-color%20or%20label%3Ainterop-2023-color',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-color',
          'interop-2023-color'
        ]
      },
      'interop-2023-container': {
        'description': 'Container Queries',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Container_Queries',
        'spec': 'https://drafts.csswg.org/css-contain-3/#container-queries',
        'tests': '/results/css/css-contain/container-queries?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-container',
        'mobile_tests': '/results/css/css-contain/container-queries?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-container',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-container'
        ]
      },
      'interop-2023-contain': {
        'description': 'Containment',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/contain',
        'spec': 'https://drafts.csswg.org/css-contain/#contain-property',
        'tests': '/results/css?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain%20or%20label%3Ainterop-2023-contain',
        'mobile_tests': '/results/css?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-contain%20or%20label%3Ainterop-2023-contain',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-contain',
          'interop-2023-contain'
        ]
      },
      'interop-2023-pseudos': {
        'description': 'CSS Pseudo-classes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Pseudo-classes',
        'spec': 'https://drafts.csswg.org/selectors/',
        'tests': '/results/css/selectors?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-pseudos',
        'mobile_tests': '/results/css/selectors?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-pseudos',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-pseudos'
        ]
      },
      'interop-2023-property': {
        'description': 'Custom Properties',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@property',
        'spec': 'https://drafts.css-houdini.org/css-properties-values-api/',
        'tests': '/results/css/css-properties-values-api?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-property',
        'mobile_tests': '/results/css/css-properties-values-api?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-property',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-property'
        ]
      },
      'interop-2023-fonts': {
        'description': 'Font Feature Detection and Palettes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/font-palette',
        'spec': 'https://drafts.csswg.org/css-fonts-4/#font-palette-prop',
        'tests': '/results/css?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-fonts',
        'mobile_tests': '/results/css?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-fonts',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-fonts'
        ]
      },
      'interop-2023-forms': {
        'description': 'Forms',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
        'spec': 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms%20or%20label%3Ainterop-2023-forms',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-forms%20or%20label%3Ainterop-2023-forms',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-forms',
          'interop-2023-forms'
        ]
      },
      'interop-2023-has': {
        'description': ':has()',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/:has',
        'spec': 'https://drafts.csswg.org/selectors-4/#relational',
        'tests': '/results/css/selectors?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-has',
        'mobile_tests': '/results/css/selectors?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-has',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-has'
        ]
      },
      'interop-2023-inert': {
        'description': 'Inert',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Global_attributes/inert',
        'spec': 'https://html.spec.whatwg.org/multipage/interaction.html#the-inert-attribute',
        'tests': '/results/inert?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-inert',
        'mobile_tests': '/results/inert?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-inert',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-inert'
        ]
      },
      'interop-2023-cssmasking': {
        'description': 'Masking',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Masking',
        'spec': 'https://drafts.fxtf.org/css-masking/',
        'tests': '/results/css/css-masking?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssmasking',
        'mobile_tests': '/results/css/css-masking?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-cssmasking',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-cssmasking'
        ]
      },
      'interop-2023-mathfunctions': {
        'description': 'CSS Math Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Functions#math_functions',
        'spec': 'https://drafts.csswg.org/css-values-4/#math',
        'tests': '/results/css/css-values?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mathfunctions',
        'mobile_tests': '/results/css/css-values?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-mathfunctions',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-mathfunctions'
        ]
      },
      'interop-2023-mediaqueries': {
        'description': 'Media Queries 4',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Media_Queries/Using_media_queries',
        'spec': 'https://www.w3.org/TR/mediaqueries-4/',
        'tests': '/results/css/mediaqueries?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mediaqueries',
        'mobile_tests': '/results/css/mediaqueries?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-mediaqueries',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-mediaqueries'
        ]
      },
      'interop-2023-modules': {
        'description': 'Modules',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules',
        'spec': 'https://tc39.es/proposal-import-assertions/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-modules',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-modules',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-modules'
        ]
      },
      'interop-2023-motion': {
        'description': 'Motion Path',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Motion_Path',
        'spec': 'https://drafts.fxtf.org/motion-1/',
        'tests': '/results/css/motion?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-motion',
        'mobile_tests': '/results/css/motion?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-motion',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-motion'
        ]
      },
      'interop-2023-offscreencanvas': {
        'description': 'Offscreen Canvas',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/OffscreenCanvas',
        'spec': 'https://html.spec.whatwg.org/multipage/canvas.html#the-offscreencanvas-interface',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-offscreencanvas',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-offscreencanvas',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-offscreencanvas'
        ]
      },
      'interop-2023-events': {
        'description': 'Pointer and Mouse Events',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/Pointer_events',
        'spec': 'https://w3c.github.io/pointerevents/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-events',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-events',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-events'
        ]
      },
      'interop-2022-scrolling': {
        'description': 'Scrolling',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/overflow',
        'spec': 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
        'tests': '/results/css?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
        'mobile_tests': '/results/css?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
        'countsTowardScore': false,
        'labels': [
          'interop-2022-scrolling'
        ]
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'mobile_tests': '/results/css/css-transforms?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'countsTowardScore': false,
        'labels': [
          'interop-2021-transforms'
        ]
      },
      'interop-2023-url': {
        'description': 'URL',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/URL',
        'spec': 'https://url.spec.whatwg.org',
        'tests': '/results/url?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-url',
        'mobile_tests': '/results/url?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-url',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-url'
        ]
      },
      'interop-2023-webcompat': {
        'description': 'Web Compat 2023',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcompat',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-webcompat',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-webcompat'
        ]
      },
      'interop-2023-webcodecs': {
        'description': 'Web Codecs (video)',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/WebCodecs_API',
        'spec': 'https://www.w3.org/TR/webcodecs/',
        'tests': '/results/webcodecs?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcodecs',
        'mobile_tests': '/results/webcodecs?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-webcodecs',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-webcodecs'
        ]
      },
      'interop-2023-webcomponents': {
        'description': 'Web Components',
        'mdn': 'https://developer.mozilla.org/docs/Web/Web_Components',
        'spec': 'https://www.w3.org/wiki/WebComponents/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcomponents',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-webcomponents',
        'countsTowardScore': false,
        'labels': [
          'interop-2023-webcomponents'
        ]
      }
    }
  },
  '2025': {
    'browsers': ['chrome_canary', 'edge', 'firefox', 'safari'],
    'mobile_browsers': ['chrome_android', 'firefox_android'],
    'table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2025-anchor-positioning',
          'interop-2025-core-web-vitals',
          'interop-2025-modules',
          'interop-2025-navigation',
          'interop-2025-backdrop-filter',
          'interop-2025-remove-mutation-events',
          'interop-2023-events',
          'interop-2024-layout',
          'interop-2025-scrollend',
          'interop-2025-storageaccess',
          'interop-2025-details',
          'interop-2025-textdecoration',
          'interop-2025-scope',
          'interop-2025-view-transitions',
          'interop-2025-webassembly',
          'interop-2025-writingmodes',
          'interop-2025-urlpattern',
          'interop-2025-webcompat',
          'interop-2025-webrtc'
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility testing',
          'Gamepad API testing',
          'Mobile testing',
          'Privacy testing',
          'WebVTT'
        ],
        'score_as_group': true
      }
    ],
    'investigation_scores': [
      {
        'name': 'Accessibility testing',
        'url': 'https://github.com/web-platform-tests/interop-accessibility',
        'scores_over_time': [
          { 'date': '2025-08-05', 'score': 300 },
          { 'date': '2025-10-07', 'score': 450 },
          { 'date': '2025-11-05', 'score': 510 },
          { 'date': '2025-12-31', 'score': 610 }
        ]
      },
      {
        'name': 'Gamepad API testing',
        'url': 'https://github.com/web-platform-tests/interop-gamepad',
        'scores_over_time': [
          { 'date': '2025-06-01', 'score': 125 },
          { 'date': '2025-08-14', 'score': 375 },
          { 'date': '2025-09-08', 'score': 415 },
          { 'date': '2025-09-22', 'score': 540 }
        ]
      },
      {
        'name': 'Mobile testing',
        'url': 'https://github.com/web-platform-tests/interop-mobile-testing',
        'scores_over_time': [
          { 'date': '2025-06-10', 'score': 120 },
          { 'date': '2025-08-26', 'score': 300 },
          { 'date': '2025-12-31', 'score': 460 }
        ]
      },
      {
        'name': 'Privacy testing',
        'url': 'https://github.com/web-platform-tests/interop-privacy',
        'scores_over_time': []
      },
      {
        'name': 'WebVTT',
        'url': 'https://github.com/web-platform-tests/interop-webvtt',
        'scores_over_time': [
          { 'date': '2025-06-23', 'score': 100 },
          { 'date': '2025-11-05', 'score': 211 }
        ]
      }
    ],
    'investigation_weight': 0.0,
    'mobile_table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2025-anchor-positioning',
          'interop-2025-core-web-vitals',
          'interop-2025-modules',
          'interop-2025-navigation',
          'interop-2025-backdrop-filter',
          'interop-2025-remove-mutation-events',
          'interop-2023-events',
          'interop-2024-layout',
          'interop-2025-scrollend',
          'interop-2025-storageaccess',
          'interop-2025-details',
          'interop-2025-textdecoration',
          'interop-2025-scope',
          'interop-2025-view-transitions',
          'interop-2025-webassembly',
          'interop-2025-writingmodes',
          'interop-2025-urlpattern',
          'interop-2025-webcompat',
          'interop-2025-webrtc'
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility testing',
          'Gamepad API testing',
          'Mobile testing',
          'Privacy testing',
          'WebVTT'
        ],
        'score_as_group': true
      }
    ],
    'mobile_focus_areas': [
      'interop-2025-anchor-positioning',
      'interop-2025-core-web-vitals',
      'interop-2025-modules',
      'interop-2025-navigation',
      'interop-2025-backdrop-filter',
      'interop-2025-remove-mutation-events',
      'interop-2023-events',
      'interop-2024-layout',
      'interop-2025-scrollend',
      'interop-2025-storageaccess',
      'interop-2025-details',
      'interop-2025-textdecoration',
      'interop-2025-scope',
      'interop-2025-view-transitions',
      'interop-2025-webassembly',
      'interop-2025-writingmodes',
      'interop-2025-urlpattern',
      'interop-2025-webcompat',
      'interop-2025-webrtc'
    ],
    /**
     * More information on results generation at
     * https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': 'https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/interop-2025/interop-2025-{stable|experimental}-v2.csv',
    'mobile_csv_url': 'https://api.github.com/repos/jgraham/interop-results/contents/2025/latest/aligned/mobile-{stable|experimental}-current.csv?ref=main',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas_description': 'https://github.com/web-platform-tests/interop/blob/main/2025/README.md',
    'focus_areas': {
      'interop-2025-anchor-positioning': {
        'description': 'CSS anchor positioning',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_anchor_positioning',
        'spec': 'https://drafts.csswg.org/css-anchor-position-1/',
        'tests': '/results/css/css-anchor-position?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-anchor-positioning',
        'mobile_tests': '/results/css/css-anchor-position?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-anchor-positioning',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-anchor-positioning'
        ]
      },
      'interop-2025-core-web-vitals': {
        'description': 'Core Web Vitals',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-core-web-vitals',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-core-web-vitals',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-core-web-vitals'
        ]
      },
      'interop-2025-scope': {
        'description': '@scope',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/@scope',
        'spec': 'https://drafts.csswg.org/css-cascade-6/#scoped-styles',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-scope',
        'mobile_tests': '/results/css/css-cascade?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-scope',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-scope'
        ]
      },
      'interop-2025-writingmodes': {
        'description': 'Writing modes',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/writing-mode',
        'spec': 'https://drafts.csswg.org/css-writing-modes/',
        'tests': '/results/css?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-writingmodes',
        'mobile_tests': '/results/css?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-writingmodes',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-writingmodes'
        ]
      },
      'interop-2024-layout': {
        'description': 'Layout',
        'mdn': '',
        'spec': '',
        'tests': '/results/css?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox%20or%20label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid%20or%20label%3Ainterop-2022-subgrid',
        'mobile_tests': '/results/css?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox%20or%20label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid%20or%20label%3Ainterop-2022-subgrid',
        'countsTowardScore': true,
        'labels': [
          'interop-2021-flexbox',
          'interop-2021-grid',
          'interop-2022-subgrid',
          'interop-2023-flexbox',
          'interop-2023-grid'
        ]
      },
      'interop-2025-modules': {
        'description': 'Modules',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/import/with',
        'spec': 'https://tc39.es/proposal-import-attributes/',
        'tests': '/results/html/semantics/scripting-1/the-script-element?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-modules',
        'mobile_tests': '/results/html/semantics/scripting-1/the-script-element?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-modules',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-modules'
        ]
      },
      'interop-2025-navigation': {
        'description': 'Navigation API',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/Navigation_API',
        'spec': 'https://html.spec.whatwg.org/multipage/nav-history-apis.html#navigation-api',
        'tests': '/results/navigation-api?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-navigation',
        'mobile_tests': '/results/navigation-api?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-navigation',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-navigation'
        ]
      },
      'interop-2025-backdrop-filter': {
        'description': 'backdrop-filter',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/backdrop-filter',
        'spec': 'https://drafts.fxtf.org/filter-effects-2/#BackdropFilterProperty',
        'tests': '/results/css/filter-effects?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-backdrop-filter',
        'mobile_tests': '/results/css/filter-effects?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-backdrop-filter',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-backdrop-filter'
        ]
      },
      'interop-2025-remove-mutation-events': {
        'description': 'Remove mutation events',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/MutationEvent',
        'spec': '',
        'tests': '/results/dom?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-remove-mutation-events',
        'mobile_tests': '/results/dom?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-remove-mutation-events',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-remove-mutation-events'
        ]
      },
      'interop-2023-events': {
        'description': 'Pointer and mouse events',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/Pointer_events',
        'spec': 'https://w3c.github.io/pointerevents/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-events',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2023-events',
        'countsTowardScore': true,
        'labels': [
          'interop-2023-events'
        ]
      },
      'interop-2025-scrollend': {
        'description': 'scrollend event',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/Document/scrollend_event',
        'spec': 'https://drafts.csswg.org/cssom-view/#eventdef-document-scrollend',
        'tests': '/results/dom/events/scrolling?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-scrollend',
        'mobile_tests': '/results/dom/events/scrolling?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-scrollend',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-scrollend'
        ]
      },
      'interop-2025-storageaccess': {
        'description': 'Storage Access API',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/Storage_Access_API',
        'spec': 'https://privacycg.github.io/storage-access/',
        'tests': '/results/storage-access-api?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-storageaccess',
        'mobile_tests': '/results/storage-access-api?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-storageaccess',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-storageaccess'
        ]
      },
      'interop-2025-details': {
        'description': '<details> element',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/HTML/Element/details',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-details-element',
        'tests': '/results/html?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-details',
        'mobile_tests': '/results/html?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-details',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-details'
        ]
      },
      'interop-2025-textdecoration': {
        'description': 'text-decoration',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/text-decoration',
        'spec': 'https://drafts.csswg.org/css-text-decor/#text-decoration-property',
        'tests': '/results/css/css-text-decor/parsing?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-textdecoration',
        'mobile_tests': '/results/css/css-text-decor/parsing?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-textdecoration',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-textdecoration'
        ]
      },
      'interop-2025-view-transitions': {
        'description': 'View Transition API',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/View_Transition_API',
        'spec': 'https://drafts.csswg.org/css-view-transitions/',
        'tests': '/results/css/css-view-transitions?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-view-transitions',
        'mobile_tests': '/results/css/css-view-transitions?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-view-transitions',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-view-transitions'
        ]
      },
      'interop-2025-webassembly': {
        'description': 'WebAssembly',
        'mdn': 'https://developer.mozilla.org/en-US/docs/WebAssembly',
        'spec': 'https://webassembly.github.io/spec/',
        'tests': '/results/wasm/jsapi?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-webassembly',
        'mobile_tests': '/results/wasm/jsapi?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-webassembly',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-webassembly'
        ]
      },
      'interop-2025-urlpattern': {
        'description': 'URLPattern',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/URL_Pattern_API',
        'spec': 'https://urlpattern.spec.whatwg.org/',
        'tests': '/results/urlpattern?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-urlpattern',
        'mobile_tests': '/results/urlpattern?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-urlpattern',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-urlpattern'
        ]
      },
      'interop-2025-webcompat': {
        'description': 'Web compat',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-webcompat',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-webcompat',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-webcompat'
        ]
      },
      'interop-2025-webrtc': {
        'description': 'WebRTC',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API',
        'spec': 'https://w3c.github.io/webrtc-pc/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2025-webrtc',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2025-webrtc',
        'countsTowardScore': true,
        'labels': [
          'interop-2025-webrtc'
        ]
      }
    }
  },
  '2026': {
    'browsers': ['chrome_canary', 'edge', 'firefox', 'safari'],
    'mobile_browsers': ['chrome_android', 'firefox_android'],
    'table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2026-anchor-positioning',
          'interop-2026-attr',
          'interop-2026-contrast-color',
          'interop-2026-container-style-queries',
          'interop-2026-custom-highlights',
          'interop-2026-dialogs-and-popovers',
          'interop-2026-fetch',
          'interop-2026-indexeddb',
          'interop-2026-jspi-for-wasm',
          'interop-2026-media-pseudo-classes',
          'interop-2026-navigation',
          'interop-2026-scoped-custom-element-registries',
          'interop-2026-scroll-driven-animations',
          'interop-2026-scroll-snap',
          'interop-2026-shape',
          'interop-2026-view-transitions',
          'interop-2026-webcompat',
          'interop-2026-webrtc',
          'interop-2026-webtransport',
          'interop-2026-zoom',
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility testing',
          'JPEG XL',
          'Mobile testing',
          'WebVTT'
        ],
        'score_as_group': true
      }
    ],
    'investigation_scores': [
      {
        'name': 'Accessibility testing',
        'url': 'https://github.com/web-platform-tests/interop-accessibility',
        'scores_over_time': []
      },
      {
        'name': 'JPEG XL',
        'url': 'https://github.com/web-platform-tests/interop-jpegxl',
        'scores_over_time': []
      },
      {
        'name': 'Mobile testing',
        'url': 'https://github.com/web-platform-tests/interop-mobile-testing',
        'scores_over_time': []
      },
      {
        'name': 'WebVTT',
        'url': 'https://github.com/web-platform-tests/interop-webvtt',
        'scores_over_time': []
      }
    ],
    'investigation_weight': 0.0,
    'mobile_table_sections': [
      {
        'name': 'Active Focus Areas',
        'rows': [
          'interop-2026-anchor-positioning',
          'interop-2026-attr',
          'interop-2026-contrast-color',
          'interop-2026-container-style-queries',
          'interop-2026-custom-highlights',
          'interop-2026-dialogs-and-popovers',
          'interop-2026-fetch',
          'interop-2026-indexeddb',
          'interop-2026-jspi-for-wasm',
          'interop-2026-media-pseudo-classes',
          'interop-2026-navigation',
          'interop-2026-scoped-custom-element-registries',
          'interop-2026-scroll-driven-animations',
          'interop-2026-scroll-snap',
          'interop-2026-shape',
          'interop-2026-view-transitions',
          'interop-2026-webcompat',
          'interop-2026-webrtc',
          'interop-2026-webtransport',
          'interop-2026-zoom',
        ],
        'score_as_group': false
      },
      {
        'name': 'Active Investigations',
        'rows': [
          'Accessibility testing',
          'JPEG XL',
          'Mobile testing',
          'WebVTT'
        ],
        'score_as_group': true
      }
    ],
    'mobile_focus_areas': [
      'interop-2026-anchor-positioning',
      'interop-2026-attr',
      'interop-2026-contrast-color',
      'interop-2026-container-style-queries',
      'interop-2026-custom-highlights',
      'interop-2026-dialogs-and-popovers',
      'interop-2026-fetch',
      'interop-2026-indexeddb',
      'interop-2026-jspi-for-wasm',
      'interop-2026-media-pseudo-classes',
      'interop-2026-navigation',
      'interop-2026-scoped-custom-element-registries',
      'interop-2026-scroll-driven-animations',
      'interop-2026-scroll-snap',
      'interop-2026-shape',
      'interop-2026-view-transitions',
      'interop-2026-webcompat',
      'interop-2026-webrtc',
      'interop-2026-webtransport',
      'interop-2026-zoom',
    ],
    /**
     * More information on results generation at
     * https://github.com/web-platform-tests/results-analysis
    **/
    'csv_url': 'https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/interop-2026/interop-2026-{stable|experimental}-v2.csv',
    'mobile_csv_url': 'https://api.github.com/repos/jgraham/interop-results/contents/2026/latest/aligned/mobile-{stable|experimental}-current.csv?ref=main',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas_description': 'https://github.com/web-platform-tests/interop/blob/main/2026/README.md',
    'focus_areas': {
      'interop-2026-anchor-positioning': {
        'description': 'CSS anchor positioning',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_anchor_positioning',
        'spec': 'https://drafts.csswg.org/css-anchor-position-1/',
        'tests': '/results/css/css-anchor-position?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-anchor-positioning',
        'mobile_tests': '/results/css/css-anchor-position?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-anchor-positioning',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-anchor-positioning'
        ]
      },
      'interop-2026-attr': {
        'description': 'CSS attr()',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-values?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-attr',
        'mobile_tests': '/results/css/css-values?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-attr',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-attr'
        ]
      },
      'interop-2026-contrast-color': {
        'description': 'CSS contrast-color()',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-color?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-contrast-color',
        'mobile_tests': '/results/css/css-color?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-contrast-color',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-contrast-color'
        ]
      },
      'interop-2026-container-style-queries': {
        'description': 'Container style queries',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-conditional/container-queries?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-container-style-queries',
        'mobile_tests': '/results/css/css-conditional/container-queries?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-container-style-queries',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-container-style-queries'
        ]
      },
      'interop-2026-custom-highlights': {
        'description': 'Custom highlights',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-highlight-api?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-custom-highlights',
        'mobile_tests': '/results/css/css-highlight-api?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-custom-highlights',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-custom-highlights'
        ]
      },
      'interop-2026-dialogs-and-popovers': {
        'description': 'Dialogs and popovers',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-dialogs-and-popovers',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-dialogs-and-popovers',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-dialogs-and-popovers'
        ]
      },
      'interop-2026-fetch': {
        'description': 'Fetch uploads and ranges',
        'mdn': '',
        'spec': '',
        'tests': '/results/fetch?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-fetch',
        'mobile_tests': '/results/fetch?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-fetch',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-fetch'
        ]
      },
      'interop-2026-indexeddb': {
        'description': 'IndexedDB',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API/Using_IndexedDB',
        'spec': '',
        'tests': '/results/IndexedDB?label=master&label=experimental&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-indexeddb',
        'mobile_tests': '/results/IndexedDB?label=master&label=experimental&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-indexeddb',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-indexeddb'
        ]
      },
      'interop-2026-jspi-for-wasm': {
        'description': 'JSPI for WASM',
        'mdn': '',
        'spec': '',
        'tests': '/results/wasm/jsapi/jspi?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-jspi-for-wasm',
        'mobile_tests': '/results/wasm/jsapi/jspi?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-jspi-for-wasm',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-jspi-for-wasm'
        ]
      },
      'interop-2026-media-pseudo-classes': {
        'description': 'Media pseudo-classes',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/selectors/media?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-media-pseudo-classes',
        'mobile_tests': '/results/css/selectors/media?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-media-pseudo-classes',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-media-pseudo-classes'
        ]
      },
      'interop-2026-navigation': {
        'description': 'Navigation API',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/Navigation_API',
        'spec': 'https://html.spec.whatwg.org/multipage/nav-history-apis.html#navigation-api',
        'tests': '/results/navigation-api?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-navigation',
        'mobile_tests': '/results/navigation-api?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-navigation',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-navigation'
        ]
      },
      'interop-2026-scoped-custom-element-registries': {
        'description': 'Scoped custom element registries',
        'mdn': '',
        'spec': '',
        'tests': '/results/custom-elements/registries?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-scoped-custom-element-registries',
        'mobile_tests': '/results/custom-elements/registries?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-scoped-custom-element-registries',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-scoped-custom-element-registries'
        ]
      },
      'interop-2026-scroll-driven-animations': {
        'description': 'Scroll-driven animations',
        'mdn': '',
        'spec': '',
        'tests': '/results/scroll-animations?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-scroll-driven-animations',
        'mobile_tests': '/results/scroll-animations?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-scroll-driven-animations',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-scroll-driven-animations'
        ]
      },
      'interop-2026-scroll-snap': {
        'description': 'Scroll snap',
        'mdn': '',
        'spec': '',
        'tests': '/results/css/css-scroll-snap?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-scroll-snap',
        'mobile_tests': '/results/css/css-scroll-snap?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-scroll-snap',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-scroll-snap'
        ]
      },
      'interop-2026-shape': {
        'description': 'CSS shape()',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-shape',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-shape',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-shape'
        ]
      },
      'interop-2026-view-transitions': {
        'description': 'View transitions',
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-view-transitions/',
        'tests': '/results/css/css-view-transitions?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-view-transitions',
        'mobile_tests': '/results/css/css-view-transitions?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-view-transitions',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-view-transitions'
        ]
      },
      'interop-2026-webcompat': {
        'description': 'Web compat',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-webcompat',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-webcompat',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-webcompat'
        ]
      },
      'interop-2026-webrtc': {
        'description': 'WebRTC',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API',
        'spec': 'https://w3c.github.io/webrtc-pc/',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-webrtc',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-webrtc',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-webrtc'
        ]
      },
      'interop-2026-webtransport': {
        'description': 'WebTransport',
        'mdn': '',
        'spec': '',
        'tests': '/results/webtransport?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-webtransport',
        'mobile_tests': '/results/webtransport?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-webtransport',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-webtransport'
        ]
      },
      'interop-2026-zoom': {
        'description': 'CSS zoom',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=edge&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2026-zoom',
        'mobile_tests': '/results/?label=master&product=chrome_android&product=firefox_android&aligned&view=interop&q=label%3Ainterop-2026-zoom',
        'countsTowardScore': true,
        'labels': [
          'interop-2026-zoom'
        ]
      }
    }
  }
};
