// This file should match the data in webapp/static/interop-data.json.
// The JSON file is used by some Mozilla infrastructure, so the file should
// not be deleted and should match the data in this file.
export const interopData = {
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
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio'
      },
      'interop-2021-flexbox': {
        'description': 'Flexbox',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox'
      },
      'interop-2021-grid': {
        'description': 'Grid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/grid',
        'spec': 'https://drafts.csswg.org/css-grid-1/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid'
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky'
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms'
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
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio'
      },
      'interop-2021-flexbox': {
        'description': 'Flexbox',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox'
      },
      'interop-2021-grid': {
        'description': 'Grid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/grid',
        'spec': 'https://drafts.csswg.org/css-grid-1/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid'
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky'
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms'
      },
      'interop-2022-cascade': {
        'description': 'Cascade Layers',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@layer',
        'spec': 'https://drafts.csswg.org/css-cascade/#layering',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade'
      },
      'interop-2022-color': {
        'description': 'Color Spaces and Functions',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/color_value',
        'spec': 'https://drafts.csswg.org/css-color/',
        'tests': '/results/css/css-color?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color'
      },
      'interop-2022-contain': {
        'countsTowardScore': true,
        'description': 'Containment',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/contain',
        'spec': 'https://drafts.csswg.org/css-contain/#contain-property',
        'tests': '/results/css/css-contain?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain'
      },
      'interop-2022-dialog': {
        'description': 'Dialog Element',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog'
      },
      'interop-2022-forms': {
        'description': 'Forms',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
        'spec': 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms'
      },
      'interop-2022-scrolling': {
        'description': 'Scrolling',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/overflow',
        'spec': 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling'
      },
      'interop-2022-subgrid': {
        'description': 'Subgrid',
        'countsTowardScore': true,
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
        'spec': 'https://drafts.csswg.org/css-grid-2/#subgrids',
        'tests': '/results/css/css-grid/subgrid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-subgrid'
      },
      'interop-2022-text': {
        'description': 'Typography and Encodings',
        'countsTowardScore': true,
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text'
      },
      'interop-2022-viewport': {
        'description': 'Viewport Units',
        'countsTowardScore': true,
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-values/#viewport-relative-units',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport'
      },
      'interop-2022-webcompat': {
        'description': 'Web Compat',
        'countsTowardScore': true,
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat'
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
        'url': 'https://github.com/web-platform-tests/interop-2023-accessibility-testing',
        'scores_over_time': [
          { 'date': '2023-03-08', 'score': 600 },
          { 'date': '2023-05-13', 'score': 700 },
          { 'date': '2023-06-27', 'score': 780 },
          { 'date': '2023-09-05', 'score': 860 },
          { 'date': '2023-09-27', 'score': 870 },
          { 'date': '2023-10-12', 'score': 910 },
          { 'date': '2023-10-13', 'score': 920 }
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
    'csv_url': 'https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/interop-2023/interop-2023-{stable|experimental}-v2.csv',
    'summary_feature_name': 'summary',
    'issue_url': 'https://github.com/web-platform-tests/interop/issues/new',
    'focus_areas': {
      'interop-2021-aspect-ratio': {
        'description': 'Aspect Ratio',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
        'spec': 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
        'countsTowardScore': false
      },
      'interop-2021-position-sticky': {
        'description': 'Sticky Positioning',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/position',
        'spec': 'https://drafts.csswg.org/css-position/#position-property',
        'tests': '/results/css/css-position/sticky?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
        'countsTowardScore': false
      },
      'interop-2022-cascade': {
        'description': 'Cascade Layers',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@layer',
        'spec': 'https://drafts.csswg.org/css-cascade/#layering',
        'tests': '/results/css/css-cascade?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade',
        'countsTowardScore': false
      },
      'interop-2022-dialog': {
        'description': 'Dialog Element',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
        'spec': 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog',
        'countsTowardScore': false
      },
      'interop-2022-text': {
        'description': 'Typography and Encodings',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/length#relative_length_units_based_on_viewport',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text',
        'countsTowardScore': false
      },
      'interop-2022-viewport': {
        'description': 'Viewport Units',
        'mdn': '',
        'spec': 'https://drafts.csswg.org/css-values/#viewport-relative-units',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport',
        'countsTowardScore': false
      },
      'interop-2022-webcompat': {
        'description': 'Web Compat',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
        'countsTowardScore': false
      },
      'interop-2023-cssborderimage': {
        'description': 'Border Image',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/border-image',
        'spec': 'https://www.w3.org/TR/css-backgrounds-3/#the-border-image',
        'tests': '/results/css/css-backgrounds?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssborderimage',
        'countsTowardScore': true
      },
      'interop-2023-color': {
        'description': 'Color Spaces and Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/color_value',
        'spec': 'https://w3c.github.io/csswg-drafts/css-color/#color-syntax',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color%20or%20label%3Ainterop-2023-color',
        'countsTowardScore': true
      },
      'interop-2023-container': {
        'description': 'Container Queries',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Container_Queries',
        'spec': 'https://drafts.csswg.org/css-contain-3/#container-queries',
        'tests': '/results/css/css-contain/container-queries?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-container',
        'countsTowardScore': true
      },
      'interop-2023-contain': {
        'description': 'Containment',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/contain',
        'spec': 'https://drafts.csswg.org/css-contain/#contain-property',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain%20or%20label%3Ainterop-2023-contain',
        'countsTowardScore': true
      },
      'interop-2023-pseudos': {
        'description': 'CSS Pseudo-classes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Pseudo-classes',
        'spec': 'https://drafts.csswg.org/selectors/',
        'tests': '/results/css/selectors?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-pseudos',
        'countsTowardScore': true
      },
      'interop-2023-property': {
        'description': 'Custom Properties',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/@property',
        'spec': 'https://drafts.css-houdini.org/css-properties-values-api/',
        'tests': '/results/css/css-properties-values-api?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-property',
        'countsTowardScore': true
      },
      'interop-2023-flexbox': {
        'description': 'Flexbox',
        'mdn': 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
        'spec': 'https://drafts.csswg.org/css-flexbox/',
        'tests': '/results/css/css-flexbox?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox%20or%20label%3Ainterop-2023-flexbox',
        'countsTowardScore': true
      },
      'interop-2023-fonts': {
        'description': 'Font Feature Detection and Palettes',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/font-palette',
        'spec': 'https://drafts.csswg.org/css-fonts-4/#font-palette-prop',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-fonts',
        'countsTowardScore': true
      },
      'interop-2023-forms': {
        'description': 'Forms',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
        'spec': 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms%20or%20label%3Ainterop-2023-forms',
        'countsTowardScore': true
      },
      'interop-2023-grid': {
        'description': 'Grid',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout',
        'spec': 'https://drafts.csswg.org/css-grid/',
        'tests': '/results/css/css-grid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid%20or%20label%3Ainterop-2023-grid',
        'countsTowardScore': true
      },
      'interop-2023-has': {
        'description': ':has()',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/:has',
        'spec': 'https://drafts.csswg.org/selectors-4/#relational',
        'tests': '/results/css/selectors?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-has',
        'countsTowardScore': true
      },
      'interop-2023-inert': {
        'description': 'Inert',
        'mdn': 'https://developer.mozilla.org/docs/Web/HTML/Global_attributes/inert',
        'spec': 'https://html.spec.whatwg.org/multipage/interaction.html#the-inert-attribute',
        'tests': '/results/inert?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-inert',
        'countsTowardScore': true
      },
      'interop-2023-cssmasking': {
        'description': 'Masking',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Masking',
        'spec': 'https://drafts.fxtf.org/css-masking/',
        'tests': '/results/css/css-masking?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-cssmasking',
        'countsTowardScore': true
      },
      'interop-2023-mathfunctions': {
        'description': 'CSS Math Functions',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Functions#math_functions',
        'spec': 'https://drafts.csswg.org/css-values-4/#math',
        'tests': '/results/css/css-values?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mathfunctions',
        'countsTowardScore': true
      },
      'interop-2023-mediaqueries': {
        'description': 'Media Queries 4',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/Media_Queries/Using_media_queries',
        'spec': 'https://www.w3.org/TR/mediaqueries-4/',
        'tests': '/results/css/mediaqueries?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-mediaqueries',
        'countsTowardScore': true
      },
      'interop-2023-modules': {
        'description': 'Modules',
        'mdn': 'https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules',
        'spec': 'https://tc39.es/proposal-import-assertions/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-modules',
        'countsTowardScore': true
      },
      'interop-2023-motion': {
        'description': 'Motion Path',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Motion_Path',
        'spec': 'https://drafts.fxtf.org/motion-1/',
        'tests': '/results/css/motion?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-motion',
        'countsTowardScore': true
      },
      'interop-2023-offscreencanvas': {
        'description': 'Offscreen Canvas',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/OffscreenCanvas',
        'spec': 'https://html.spec.whatwg.org/multipage/canvas.html#the-offscreencanvas-interface',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-offscreencanvas',
        'countsTowardScore': true
      },
      'interop-2023-events': {
        'description': 'Pointer and Mouse Events',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/Pointer_events',
        'spec': 'https://w3c.github.io/pointerevents/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-events',
        'countsTowardScore': true
      },
      'interop-2022-scrolling': {
        'description': 'Scrolling',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/overflow',
        'spec': 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
        'tests': '/results/css?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
        'countsTowardScore': true
      },
      'interop-2022-subgrid': {
        'description': 'Subgrid',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
        'spec': 'https://drafts.csswg.org/css-grid-2/#subgrids',
        'tests': '/results/css/css-grid/subgrid?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-subgrid',
        'countsTowardScore': true
      },
      'interop-2021-transforms': {
        'description': 'Transforms',
        'mdn': 'https://developer.mozilla.org/docs/Web/CSS/transform',
        'spec': 'https://drafts.csswg.org/css-transforms/',
        'tests': '/results/css/css-transforms?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
        'countsTowardScore': true
      },
      'interop-2023-url': {
        'description': 'URL',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/URL',
        'spec': 'https://url.spec.whatwg.org',
        'tests': '/results/url?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-url',
        'countsTowardScore': true
      },
      'interop-2023-webcompat': {
        'description': 'Web Compat 2023',
        'mdn': '',
        'spec': '',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcompat',
        'countsTowardScore': true
      },
      'interop-2023-webcodecs': {
        'description': 'Web Codecs (video)',
        'mdn': 'https://developer.mozilla.org/docs/Web/API/WebCodecs_API',
        'spec': 'https://www.w3.org/TR/webcodecs/',
        'tests': '/results/webcodecs?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcodecs',
        'countsTowardScore': true
      },
      'interop-2023-webcomponents': {
        'description': 'Web Components',
        'mdn': 'https://developer.mozilla.org/docs/Web/Web_Components',
        'spec': 'https://www.w3.org/wiki/WebComponents/',
        'tests': '/results/?label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2023-webcomponents',
        'countsTowardScore': true
      }
    }
  }
};
