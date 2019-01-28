/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-tooltip/paper-tooltip.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { WPTFlags } from './wpt-flags.js';
import './ohm.js';

/* eslint-enable */

/* global ohm */
const QUERY_GRAMMAR = ohm.grammar(`
  Query {
    Q = ListOf<Exp, space*>

    Exp = NonemptyListOf<OrPart, or>

    NestedExp
      = "(" Exp ")"   -- paren
      | not NestedExp -- not

    OrPart = NonemptyListOf<AndPart, and>

    AndPart
      = NestedExp
      | Fragment    -- fragment

    or
      = "|"
      | caseInsensitive<"or">

    and
      = "&"
      | caseInsensitive<"and">

    not
      = "!"
      | "not"

    Fragment
      = not Fragment -- not
      | nameLiteral
      | statusExp
      | nameFragment

    nameLiteral
      = "\\"" nameLiteralInner "\\""

    nameLiteralInner
      = nameLiteralChar*

    nameLiteralChar
      = "\\x00".."\\x21"
      | "\\x5D".."\\uFFFF"
      | "\\\\" -- slash
      | "\\"" -- quote

    statusExp
      = caseInsensitive<"status"> ":" statusLiteral  -- eq
      | caseInsensitive<"status"> ":!" statusLiteral -- neq
      | browserName ":" statusLiteral                -- browser_eq
      | browserName ":!" statusLiteral               -- browser_neq

    browserName
      = caseInsensitive<"chrome">
      | caseInsensitive<"edge">
      | caseInsensitive<"firefox">
      | caseInsensitive<"safari">

    statusLiteral
      = caseInsensitive<"unknown">
      | caseInsensitive<"pass">
      | caseInsensitive<"ok">
      | caseInsensitive<"error">
      | caseInsensitive<"timeout">
      | caseInsensitive<"notrun">
      | caseInsensitive<"fail">
      | caseInsensitive<"crash">
      | caseInsensitive<"skip">
      | caseInsensitive<"assert">

    nameFragment = nameFragmentChar+

    nameFragmentChar
      = "\\x00".."\\x08"
      | "\\x0E".."\\x1F"
      | "\\x21".."\\uFFFF"
  }
`);
/* eslint-disable */
const evalNot = (n, p) => {
  return {not: p.eval()};
};
const evalSelf = p => p.eval();
const QUERY_SEMANTICS = QUERY_GRAMMAR.createSemantics().addOperation('eval', {
  _terminal: function() {
    return this.sourceString;
  },
  EmptyListOf: function() {
    return [];
  },
  NonemptyListOf: function(fst, seps, rest) {
    return [fst.eval()].concat(rest.eval());
  },
  Q: l => {
    const ps = l.eval();
    // Separate atoms are each treated as "there exists a run where ...",
    // and the root is grouped by AND of the separated atoms.
    // Nested ands, on the other hand, require all conditions to be met by the same run.
    return ps.length === 0
      ? {exists: [{pattern: ''}]}
      : {exists: ps };
  },
  Exp: l => {
    const ps = l.eval();
    return ps.length === 1 ? ps[0] : {or: ps};
  },
  NestedExp: evalSelf,
  NestedExp_paren: (_, p, __) => p.eval(),
  NestedExp_not: evalNot,
  OrPart: l => {
    const ps = l.eval();
    return ps.length === 1 ? ps[0] : {and: ps};
  },
  AndPart_fragment: evalSelf,
  Fragment: evalSelf,
  Fragment_not: evalNot,
  nameLiteral: (_, l, __) => {
    return {pattern: l.eval().join('')};
  },
  nameLiteralInner: chars => chars.eval(),
  nameLiteralChar_slash: (v) => '\\',
  nameLiteralChar_quote: (v) => '"',
  statusExp_eq: (l, colon, r) => {
    return { status: r.sourceString.toUpperCase() };
  },
  statusExp_browser_eq: (l, colon, r) => {
    return {
      browser_name: l.sourceString.toLowerCase(),
      status: r.sourceString.toUpperCase(),
    };
  },
  statusExp_neq: (l, colonBang, r) => {
    return { status: {not: r.sourceString.toUpperCase() } };
  },
  statusExp_browser_neq: (l, colonBang, r) => {
    return {
      browser_name: l.sourceString.toLowerCase(),
      status: {not: r.sourceString.toUpperCase()},
    };
  },
  nameFragment: (chars) => {
    return {pattern: chars.eval().join('')};
  },
});
/* eslint-enable */

const QUERY_DEBOUNCE_ID = Symbol('query_debounce_timeout');

class TestSearch extends WPTFlags(PolymerElement) {
  static get template() {
    return html`
    <style>
      input.query {
        font-size: 16px;
        display: block;
        padding: 0.5em 0;
        width: 100%;
      }
    </style>

    <div>
      <input value="{{ queryInput::input }}" class="query" list="query-list" placeholder="[[queryPlaceholder]]" onchange="[[onChange]]" onkeyup="[[onKeyUp]]" onkeydown="[[onKeyDown]]" onfocus="[[onFocus]]" onblur="[[onBlur]]">
      <!-- TODO(markdittmer): Static id will break multiple search
        components. -->
      <datalist id="query-list"></datalist>
      <paper-tooltip position="top" manual-mode="true">
        Press &lt;Enter&gt; to commit query
      </paper-tooltip>
    </div>
`;
  }

  static get QUERY_GRAMMAR() {
    return QUERY_GRAMMAR;
  }
  static get QUERY_SEMANTICS() {
    return QUERY_SEMANTICS;
  }
  static get is() {
    return 'test-search';
  }
  static get properties() {
    return {
      placeholder: {
        type: String,
        value: 'Search test files, like \'cors/allow-headers.htm\', then press <Enter>',
      },
      // Query input string
      queryInput: {
        type: String,
        notify: true,
        observer: 'queryInputChanged'
      },
      // Debounced + normalized query string.
      query: {
        type: String,
        notify: true,
        observer: 'queryUpdated',
      },
      structuredQuery: Object,
      results: {
        type: Array,
        notify: true,
      },
      queryPlaceholder: {
        type: String,
        computed: 'computeQueryPlaceholder()'
      },
      testPaths: Array,
      onKeyUp: Function,
      onChange: Function,
      onFocus: Function,
      onBlur: Function,
    };
  }

  constructor() {
    super();

    this.onChange = this.handleChange.bind(this);
    this.onFocus = this.handleFocus.bind(this);
    this.onBlur = this.handleBlur.bind(this);
    this.onKeyUp = this.handleKeyUp.bind(this);
  }

  ready() {
    super.ready();
    this._createMethodObserver('updateDatalist(query, testPaths)');
    this.queryInput = this.query;
  }

  queryUpdated(query) {
    this.queryInput = query;
    if (this.structuredQueries) {
      try {
        this.structuredQuery = this.parseAndInterpretQuery(query);
      } catch (err) {
        // TODO: Handle query parse/interpret error.
      }
    }
  }

  parseAndInterpretQuery(query) {
    const p = QUERY_GRAMMAR.match(query);
    if (!p.succeeded()) {
      throw new Error(`Failed to parse query: ${query}`);
    }

    return QUERY_SEMANTICS(p).eval();
  }

  updateDatalist(query, paths) {
    if (!paths) {
      return;
    }
    const datalist = this.shadowRoot.querySelector('datalist');
    datalist.innerHTML = '';
    let matches = Array.from(paths);
    if (query) {
      matches = matches
        .filter(p => p.toLowerCase())
        .filter(p => p.includes(query))
        .sort((p1, p2) => p1.indexOf(query) - p2.indexOf(query));
    }
    for (const match of matches.slice(0, 10)) {
      const option = document.createElement('option');
      option.setAttribute('value', match);
      datalist.appendChild(option);
    }
  }

  queryInputChanged(_, oldQuery) {
    // Debounce first initialization.
    if (typeof(oldQuery) === 'undefined') {
      return;
    }
    if (this[QUERY_DEBOUNCE_ID]) {
      window.clearTimeout(this[QUERY_DEBOUNCE_ID]);
    }
    this[QUERY_DEBOUNCE_ID] = window.setTimeout(this.latchQuery.bind(this), 500);
  }

  latchQuery() {
    this.query = (this.queryInput || '').toLowerCase();
  }

  commitQuery() {
    this.query = this.queryInput;
    this.dispatchEvent(new CustomEvent('commit', {
      detail: {
        query: this.query,
        structuredQuery: this.structuredQuery,
      },
    }));
    this.shadowRoot.querySelector('.query').blur();
  }

  handleKeyUp(e) {
    if (e.keyCode !== 13) {
      return;
    }

    this.commitQuery();
  }

  handleChange(e) {
    const opts = Array.from(this.shadowRoot.querySelectorAll('option'))
      .map(elem => elem.getAttribute('value').toLowerCase());
    if (opts.length === 0) {
      return;
    }

    const path = e.target.value;
    if (opts.includes(path.toLowerCase())) {
      this.dispatchEvent(new CustomEvent('autocomplete', {
        detail: {path: path},
      }));
      this.shadowRoot.querySelector('.query').blur();
    }
  }

  handleFocus() {
    this.shadowRoot.querySelector('paper-tooltip').show();
  }

  handleBlur() {
    this.shadowRoot.querySelector('paper-tooltip').hide();
  }

  clear() {
    this.query = '';
    this.queryInput = '';
  }
}
window.customElements.define(TestSearch.is, TestSearch);

export { TestSearch };
