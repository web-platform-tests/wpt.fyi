/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before interop-data-manager.js; see
// that file.
import { charts } from './util/fake-google-charts.js';
import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import { interopData } from '../interop-data.js';

// The dashboard reads the showMobileScoresView flag from localStorage once,
// when its class is first used, so set it before the elements are loaded.
const MOBILE_FLAG = 'features.showMobileScoresView';
localStorage.setItem(MOBILE_FLAG, 'true');
const { InteropDashboard } = await import('../interop-dashboard.js');
// Register the elements the way the interop page does.
await import('../interop.js');

// Most tests use 2023, whose scores are frozen: its CSVs are checked in to
// webapp/static. The expected scores were computed independently from them.
const YEAR = '2023';
const CURRENT_YEAR = [...interopData.valid_years].sort().pop();

// 2023's active focus areas, sorted by name.
const ACTIVE_FOCUS_AREAS_BY_NAME = [
  'Border Image', 'Color Spaces and Functions', 'Container Queries',
  'Containment', 'CSS Math Functions', 'CSS Pseudo-classes',
  'Custom Properties', 'Flexbox', 'Font Feature Detection and Palettes',
  'Forms', 'Grid', ':has()', 'Inert', 'Masking', 'Media Queries 4', 'Modules',
  'Motion Path', 'Offscreen Canvas', 'Pointer and Mouse Events', 'Scrolling',
  'Subgrid', 'Transforms', 'URL', 'Web Codecs (video)', 'Web Compat 2023',
  'Web Components',
];

function shadow(el) {
  return el.shadowRoot;
}

function scoreTables(el) {
  return [...shadow(el).querySelectorAll('table.score-table')];
}

// The text of each cell in each body row of a score table.
function bodyRows(table) {
  return [...table.querySelectorAll('tbody tr')]
    .map(row => [...row.querySelectorAll('td')].map(cell => cell.textContent.trim()));
}

function totalRow(table) {
  return [...table.querySelectorAll('tfoot tr td')].map(cell => cell.textContent.trim());
}

function sortableHeaders(el) {
  return [...shadow(el).querySelectorAll('.sortable-header')];
}

function sortIcons(el) {
  return sortableHeaders(el).map(header =>
    header.querySelector('img.sort-icon, img.sort-icon-focus-areas').getAttribute('src')
      .replace('/static/', ''));
}

function isDisplayed(element) {
  return getComputedStyle(element).display !== 'none';
}

// The query the dashboard last put in the URL. It replaces the whole query.
function urlQuery() {
  return window.location.search.replace(/^\?/, '');
}

// Builds a desktop CSV for a year whose data isn't checked in, with every
// score set to `score`.
function desktopCsv(year, score) {
  const yearInfo = interopData[year];
  const focusAreas = Object.keys(yearInfo.focus_areas);
  const browsers = [...yearInfo.browsers, 'interop'];
  const header = ['date'];
  const row = [`${year}-03-01`];
  for (const browser of browsers) {
    header.push(`${browser}-version`, ...focusAreas.map(f => `${browser}-${f}`));
    row.push('1.0', ...focusAreas.map(() => String(score)));
  }
  return `${header.join(',')}\n${row.join(',')}\n`;
}

suite('<interop-dashboard>', () => {
  let sandbox, originalUrl;

  setup(() => {
    sandbox = sinon.createSandbox();
    originalUrl = window.location.href;
    // Changing views would navigate away from the test page.
    sandbox.stub(InteropDashboard.prototype, 'toggleMobileView');
  });

  teardown(() => {
    sandbox.restore();
    window.history.replaceState(null, '', originalUrl);
  });

  suiteTeardown(() => {
    localStorage.removeItem(MOBILE_FLAG);
  });

  // Creates a dashboard for the year with the given URL query, e.g.
  // 'stable&feature=interop-2023-grid', and waits until it has rendered.
  async function createDashboard(year = YEAR, query = '') {
    const params = new URLSearchParams(window.location.search);
    for (const [key, value] of new URLSearchParams(query)) {
      params.append(key, value);
    }
    window.history.replaceState(null, '', `${window.location.pathname}?${params}`);
    const el = await fixture(html`<interop-dashboard year="${year}"></interop-dashboard>`);
    // ready() loads the scores before rendering anything.
    await waitUntil(
      () => shadow(el) && shadow(el).querySelectorAll('table.score-table tbody tr').length > 0,
      'expected the score tables to render', { timeout: 5000 });
    // The sort click handlers are added after the next render.
    await nextFrame();
    await nextFrame();
    return el;
  }

  test('is a registered custom element', () => {
    assert.isTrue(new InteropDashboard() instanceof HTMLElement);
    assert.isTrue(document.createElement('interop-dashboard') instanceof InteropDashboard);
  });

  suite('header', () => {
    test('shows the year in the title', async() => {
      const el = await createDashboard();
      assert.equal(shadow(el).querySelector('h1').textContent.trim(), 'Interop 2023 Dashboard');
    });

    test('links to the current year from a previous year', async() => {
      const el = await createDashboard();
      const banner = shadow(el).querySelector('.previous-year-banner');
      assert.isFalse(banner.hidden);
      assert.equal(banner.querySelector('a').getAttribute('href'), `/interop-${CURRENT_YEAR}`);
    });

    test('has no previous-year banner for the current year', async() => {
      const realFetch = window.fetch.bind(window);
      const csv = desktopCsv(CURRENT_YEAR, 500);
      sandbox.stub(window, 'fetch').callsFake((url, ...args) =>
        String(url).startsWith('/') ? realFetch(url, ...args) : Promise.resolve(new Response(csv)));
      const el = await createDashboard(CURRENT_YEAR);
      assert.isTrue(shadow(el).querySelector('.previous-year-banner').hidden);
    });

    test('links to every year in the footer', async() => {
      const el = await createDashboard();
      const links = [...shadow(el).querySelectorAll('footer a')];
      assert.deepEqual(
        links.map(a => a.getAttribute('href')),
        [...interopData.valid_years].sort().map(year => `interop-${year}`));
    });

    test('sorts a copy of the years, not the shared data', async() => {
      const el = await createDashboard();
      const years = ['2024', '2021', '2023'];
      el.dataManager.getYearProp = prop => prop === 'validYears' ? years : undefined;
      assert.deepEqual(el.getAllYears(), ['2021', '2023', '2024']);
      assert.deepEqual(years, ['2024', '2021', '2023']);
    });
  });

  suite('focus area scores', () => {
    test('lists the active focus areas by name, with each browser\'s and the Interop score', async() => {
      const el = await createDashboard();
      const rows = bodyRows(scoreTables(el)[0]);
      assert.deepEqual(rows.map(row => row[0]), ACTIVE_FOCUS_AREAS_BY_NAME);
      // Name, Chrome/Edge, Firefox, Safari, Interop.
      assert.deepEqual(rows[0], ['Border Image', '98.6%', '97.3%', '100%', '96.0%']);
      assert.deepEqual(rows[4], ['CSS Math Functions', '99.7%', '99.8%', '99.4%', '99.4%']);
    });

    test('totals each column', async() => {
      const el = await createDashboard();
      assert.deepEqual(totalRow(scoreTables(el)[0]), ['TOTAL', '99.0%', '98.1%', '97.5%', '95.5%']);
    });

    test('links each focus area to its experimental test results', async() => {
      const el = await createDashboard();
      const link = scoreTables(el)[0].querySelector('tbody tr a');
      const url = new URL(link.href);
      assert.equal(url.pathname, '/results/css/css-backgrounds');
      assert.deepEqual(url.searchParams.getAll('label'), ['master', 'experimental']);
      assert.equal(url.searchParams.get('q'), 'label:interop-2023-cssborderimage');
    });

    test('shows the previous focus areas in their own table', async() => {
      const el = await createDashboard();
      const tables = scoreTables(el);
      assert.lengthOf(tables, 4);
      assert.equal(tables[2].querySelector('thead th').textContent.trim(), 'Previous Focus Areas');
      const rows = bodyRows(tables[2]);
      assert.lengthOf(rows, 7);
      assert.deepEqual(rows[0], ['Aspect Ratio', '98.7%', '99.3%', '100%', '98.0%']);
    });

    test('shows the stable scores with ?stable', async() => {
      const el = await createDashboard(YEAR, 'stable');
      const table = scoreTables(el)[0];
      assert.deepEqual(bodyRows(table)[0], ['Border Image', '100%', '97.3%', '100%', '97.3%']);
      assert.deepEqual(totalRow(table), ['TOTAL', '97.3%', '86.4%', '95.9%', '83.3%']);
      const url = new URL(table.querySelector('tbody tr a').href);
      assert.deepEqual(url.searchParams.getAll('label'), ['master', 'stable']);
    });
  });

  suite('investigations', () => {
    test('show each investigation\'s latest progress, linked to its repository', async() => {
      const el = await createDashboard();
      const table = scoreTables(el)[1];
      assert.equal(table.querySelector('thead th').textContent.trim(), 'Active Investigations');
      assert.deepEqual(bodyRows(table), [
        ['Accessibility Testing', '100.0%'],
        ['Mobile Testing', '70.0%'],
      ]);
      assert.deepEqual(
        [...table.querySelectorAll('tbody a')].map(a => a.getAttribute('href')),
        ['https://github.com/web-platform-tests/interop-accessibility',
          'https://github.com/web-platform-tests/interop-2023-mobile-testing']);
      assert.deepEqual(totalRow(table), ['TOTAL', '85.0%']);
    });

    test('show the previous year\'s investigations', async() => {
      const el = await createDashboard();
      const table = scoreTables(el)[3];
      assert.equal(table.querySelector('thead th').textContent.trim(), 'Previous Investigations');
      assert.deepEqual(bodyRows(table), [
        ['Editing, contenteditable, and execCommand', '52.0%'],
        ['Pointer and Mouse Events', '100.0%'],
        ['Viewport Measurement', '90.0%'],
      ]);
      assert.deepEqual(totalRow(table), ['TOTAL', '80.7%']);
    });
  });

  suite('sorting', () => {
    function chromeScores(el) {
      return bodyRows(scoreTables(el)[0]).map(row => parseFloat(row[1]));
    }

    test('sorts by name by default', async() => {
      const el = await createDashboard();
      // Focus area, Chrome/Edge, Firefox, Safari, Interop.
      assert.deepEqual(sortIcons(el), [
        'expand_less.svg', 'expand_inactive.svg', 'expand_inactive.svg',
        'expand_inactive.svg', 'expand_inactive.svg']);
    });

    test('clicking a browser column sorts by its score, highest first, then lowest first', async() => {
      const el = await createDashboard();
      sortableHeaders(el)[1].click();
      await nextFrame();
      const descending = chromeScores(el);
      assert.deepEqual(descending, [...descending].sort((a, b) => b - a));
      assert.equal(sortIcons(el)[1], 'expand_more.svg');
      assert.equal(sortIcons(el)[0], 'expand_inactive.svg');

      sortableHeaders(el)[1].click();
      await nextFrame();
      const ascending = chromeScores(el);
      assert.deepEqual(ascending, [...ascending].sort((a, b) => a - b));
      // URL has the lowest Chrome/Edge score, 86.8%.
      assert.deepEqual(bodyRows(scoreTables(el)[0])[0].slice(0, 2), ['URL', '86.8%']);
      assert.equal(sortIcons(el)[1], 'expand_less.svg');
    });

    test('keeps a score column sorted after switching between Stable and Experimental', async() => {
      const el = await createDashboard();
      sortableHeaders(el)[1].click();
      await nextFrame();
      el.shadowRoot.querySelector('#toggleStable').click();
      await nextFrame();
      const stable = chromeScores(el);
      assert.deepEqual(stable, [...stable].sort((a, b) => b - a));

      el.shadowRoot.querySelector('#toggleExperimental').click();
      await nextFrame();
      const experimental = chromeScores(el);
      assert.deepEqual(experimental, [...experimental].sort((a, b) => b - a));
    });

    test('clicking the Interop column sorts by the Interop score', async() => {
      const el = await createDashboard();
      sortableHeaders(el)[4].click();
      await nextFrame();
      sortableHeaders(el)[4].click();
      await nextFrame();
      const rows = bodyRows(scoreTables(el)[0]);
      const interopScores = rows.map(row => parseFloat(row[4]));
      assert.deepEqual(interopScores, [...interopScores].sort((a, b) => a - b));
      assert.deepEqual([rows[0][0], rows[0][4]], ['Grid', '84.2%']);
      assert.equal(sortIcons(el)[4], 'expand_less.svg');
    });

    test('clicking the name column again sorts by name in reverse', async() => {
      const el = await createDashboard();
      sortableHeaders(el)[0].click();
      await nextFrame();
      const rows = bodyRows(scoreTables(el)[0]);
      assert.deepEqual(rows.map(row => row[0]), [...ACTIVE_FOCUS_AREAS_BY_NAME].reverse());
      assert.equal(sortIcons(el)[0], 'expand_more.svg');
    });

    test('only sorts the first table', async() => {
      const el = await createDashboard();
      const before = bodyRows(scoreTables(el)[2]).map(row => row[0]);
      sortableHeaders(el)[1].click();
      await nextFrame();
      assert.deepEqual(bodyRows(scoreTables(el)[2]).map(row => row[0]), before);
    });
  });

  suite('stable and experimental', () => {
    function toggles(el) {
      return ['toggleStable', 'toggleExperimental'].map(id => shadow(el).querySelector(`#${id}`));
    }

    test('shows experimental scores by default', async() => {
      const el = await createDashboard();
      const [stable, experimental] = toggles(el);
      assert.isFalse(el.stable);
      assert.isTrue(stable.classList.contains('unselected'));
      assert.isTrue(experimental.classList.contains('selected'));
      assert.equal(stable.getAttribute('aria-pressed'), 'false');
      assert.equal(experimental.getAttribute('aria-pressed'), 'true');
    });

    test('clicking Stable shows the stable scores and adds ?stable to the URL', async() => {
      const el = await createDashboard();
      const [stable, experimental] = toggles(el);
      stable.click();
      await nextFrame();
      assert.isTrue(el.stable);
      assert.isTrue(stable.classList.contains('selected'));
      assert.isTrue(experimental.classList.contains('unselected'));
      assert.equal(stable.getAttribute('aria-pressed'), 'true');
      assert.equal(experimental.getAttribute('aria-pressed'), 'false');
      assert.equal(urlQuery(), 'stable');
      assert.deepEqual(totalRow(scoreTables(el)[0]), ['TOTAL', '97.3%', '86.4%', '95.9%', '83.3%']);
    });

    test('clicking Experimental switches back', async() => {
      const el = await createDashboard(YEAR, 'stable');
      const [stable, experimental] = toggles(el);
      experimental.click();
      await nextFrame();
      assert.isFalse(el.stable);
      assert.equal(stable.getAttribute('aria-pressed'), 'false');
      assert.equal(experimental.getAttribute('aria-pressed'), 'true');
      assert.equal(urlQuery(), '');
      assert.deepEqual(totalRow(scoreTables(el)[0]), ['TOTAL', '99.0%', '98.1%', '97.5%', '95.5%']);
    });
  });

  suite('feature chart', () => {
    function featureSelect(el) {
      return shadow(el).querySelector('#featureSelect');
    }

    // The visible links. A focus area without a URL shows plain text instead,
    // and <dom-if> hides the elements it no longer shows rather than removing
    // them.
    function featureLinks(el) {
      return [...shadow(el).querySelectorAll('#featureReferenceList a, #featureReferenceList span')]
        .filter(isDisplayed)
        .map(link => [link.textContent.trim(), link.getAttribute('href')]);
    }

    function chartFeature(el) {
      return shadow(el).querySelector('interop-feature-chart').feature;
    }

    test('shows the summary by default', async() => {
      const el = await createDashboard();
      const select = featureSelect(el);
      assert.equal(select.value, 'summary');
      assert.equal(select.options[0].textContent.trim(), 'All Focus Areas');
      assert.equal(chartFeature(el), 'summary');
      assert.equal(urlQuery(), '');
    });

    test('lists the focus areas by table, without the investigations', async() => {
      const el = await createDashboard();
      const groups = [...featureSelect(el).querySelectorAll('optgroup')];
      assert.deepEqual(groups.map(g => g.label), ['Active Focus Areas', 'Previous Focus Areas']);
      assert.lengthOf(groups[0].querySelectorAll('option'), 26);
      assert.lengthOf(groups[1].querySelectorAll('option'), 7);
    });

    test('choosing a focus area charts it, links to it and adds it to the URL', async() => {
      const el = await createDashboard();
      const select = featureSelect(el);
      select.value = 'interop-2023-cssborderimage';
      select.dispatchEvent(new Event('change'));
      await nextFrame();
      assert.equal(el.feature, 'interop-2023-cssborderimage');
      assert.equal(chartFeature(el), 'interop-2023-cssborderimage');
      assert.equal(urlQuery(), 'feature=interop-2023-cssborderimage');
      const links = featureLinks(el);
      assert.deepEqual(links.map(([text]) => text), ['Spec', 'MDN', 'Tests']);
      assert.equal(links[0][1], 'https://www.w3.org/TR/css-backgrounds-3/#the-border-image');
      assert.equal(links[1][1], 'https://developer.mozilla.org/docs/Web/CSS/border-image');
      assert.deepEqual(new URL(links[2][1]).searchParams.getAll('label'), ['master', 'experimental']);
      await waitUntil(() => charts.some(chart =>
        chart.container === shadow(el).querySelector('interop-feature-chart').$.failuresChart &&
        chart.draws.length), 'expected a chart');
    });

    test('reads the feature, stable and embedded options from the URL', async() => {
      const el = await createDashboard(YEAR, 'feature=interop-2023-grid&stable&embedded');
      assert.equal(featureSelect(el).value, 'interop-2023-grid');
      assert.equal(el.feature, 'interop-2023-grid');
      assert.isTrue(el.stable);
      assert.isTrue(el.embedded);
      assert.equal(urlQuery(), 'feature=interop-2023-grid&stable&embedded');
      const tests = featureLinks(el).find(([text]) => text === 'Tests');
      assert.deepEqual(new URL(tests[1]).searchParams.getAll('label'), ['master', 'stable']);
    });
  });

  suite('formatTestsURL()', () => {
    test('replaces the stable or experimental label and keeps the others', async() => {
      const el = await createDashboard();
      const url = new URL(el.formatTestsURL('/results/?label=master&label=stable&q=x', false));
      assert.equal(url.pathname, '/results/');
      assert.deepEqual(url.searchParams.getAll('label'), ['master', 'experimental']);
      assert.equal(url.searchParams.get('q'), 'x');
    });

    test('returns an empty string for an empty URL', async() => {
      const el = await createDashboard();
      assert.strictEqual(el.formatTestsURL(undefined, true), '');
      assert.strictEqual(el.formatTestsURL('', true), '');
    });
  });

  suite('layout', () => {
    // The stylesheet switches to a single column on narrow screens, so these
    // tests check the styles that ready() sets on the elements.
    test('keeps the stylesheet\'s layout and the extra description from 2023', async() => {
      const el = await createDashboard();
      assert.equal(shadow(el).querySelector('.grid-container').style.display, '');
      assert.equal(shadow(el).querySelector('.extra-description').style.display, '');
    });

    test('uses a narrow single column without the extra description for 2022', async() => {
      const el = await createDashboard('2022');
      const grid = shadow(el).querySelector('.grid-container');
      assert.equal(grid.style.display, 'block');
      assert.equal(grid.style.maxWidth, '700px');
      assert.equal(shadow(el).querySelector('.extra-description').style.display, 'none');
    });

    test('uses a narrow single column in the mobile view', async() => {
      const el = await createDashboard('2024', 'mobile-view');
      assert.equal(shadow(el).querySelector('.grid-container').style.display, 'block');
    });
  });

  suite('mobile view', () => {
    // These tests run with the showMobileScoresView flag on.
    function mobileButton(el) {
      return shadow(el).querySelector('#toggleMobile');
    }

    test('offers a Mobile button for years with mobile scores', async() => {
      const el = await createDashboard('2024');
      assert.isFalse(mobileButton(el).hidden);
      assert.isTrue(mobileButton(el).classList.contains('unselected'));
    });

    test('has no Mobile button for years without mobile scores', async() => {
      const el = await createDashboard();
      assert.isTrue(mobileButton(el).hidden);
    });

    test('clicking Mobile switches to the mobile view', async() => {
      const el = await createDashboard('2024');
      mobileButton(el).click();
      assert.isTrue(InteropDashboard.prototype.toggleMobileView.calledOnceWith(true, false));
    });

    test('shows the mobile scores with ?mobile-view', async() => {
      const el = await createDashboard('2024', 'mobile-view');
      assert.isTrue(el.isMobileScoresView);
      assert.isFalse(el.stable);
      assert.equal(shadow(el).querySelector('h1').textContent.trim(), 'Interop 2024 Mobile Dashboard');
      assert.isFalse(shadow(el).querySelector('#mobileWarning').hidden);
      assert.isTrue(mobileButton(el).classList.contains('selected'));
      // Name, Chrome, Firefox, Safari iOS (no scores yet), Interop.
      const rows = bodyRows(scoreTables(el)[0]);
      assert.lengthOf(rows[0], 5);
      assert.equal(rows[0][3], '--%');
      assert.equal(urlQuery(), 'mobile-view');
    });

    test('ignores ?stable in the mobile view', async() => {
      const el = await createDashboard('2024', 'mobile-view&stable');
      assert.isFalse(el.stable);
      assert.equal(urlQuery(), 'mobile-view');
    });

    test('clicking Stable or Experimental leaves the mobile view', async() => {
      const el = await createDashboard('2024', 'mobile-view');
      shadow(el).querySelector('#toggleStable').click();
      shadow(el).querySelector('#toggleExperimental').click();
      const calls = InteropDashboard.prototype.toggleMobileView.getCalls().map(call => call.args);
      assert.deepEqual(calls, [[false, true], [false, false]]);
    });
  });
});
