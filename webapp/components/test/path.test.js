import { assert, expect, fixture, fixtureSync, html } from '@open-wc/testing';
import { LitElement } from 'lit';
import sinon from 'sinon';
import { html as polymerHtml, PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { LitPathInfo, PathInfo, PathPart } from '../path.js';

class ConcreteType extends PathInfo(PolymerElement) {}
window.customElements.define('path-info-concrete', ConcreteType);

suite('PathInfo', () => {
  let pathInfo;

  setup(() => {
    pathInfo = fixtureSync(html`<path-info-concrete></path-info-concrete>`);
  });

  test('pathIsATestPath', () => {
    pathInfo.path = '/a/b.html';
    expect(pathInfo.pathIsATestFile).to.be.true;
    pathInfo.path = '/a/b.html?x=1';
    expect(pathInfo.pathIsATestFile).to.be.true;
    pathInfo.path = '/a/b.html#x=1';
    expect(pathInfo.pathIsATestFile).to.be.true;
    pathInfo.path = '/a/b.html?x=1#y=2';
    expect(pathInfo.pathIsATestFile).to.be.true;
    pathInfo.path = '/a/b.html?x=:a/b/:';
    expect(pathInfo.pathIsATestFile).to.be.true;
    pathInfo.path = '/a/b/';
    expect(pathInfo.pathIsATestFile).to.be.false;
  });

  test('encodeTestPath', () => {
    expect(pathInfo.encodeTestPath('/a/b.html')).to.equal('/a/b.html');
    expect(pathInfo.encodeTestPath('/a/b/')).to.equal('/a/b/');
    expect(pathInfo.encodeTestPath('/a/b/?c=d')).to.equal('/a/b/%3Fc%3Dd');
    expect(pathInfo.encodeTestPath('/a/b/#c=d')).to.equal('/a/b/%23c%3Dd');
    expect(pathInfo.encodeTestPath('/a/b/?c=d#e=f')).to.equal('/a/b/%3Fc%3Dd%23e%3Df');

    // Specific example.
    expect(pathInfo.encodeTestPath('/webgpu/cts.html?q=cts:command_buffer/basic:'))
      .to.equal('/webgpu/cts.html%3Fq%3Dcts%3Acommand_buffer%2Fbasic%3A');
  });

  test('isParentDir', () => {
    assert.isTrue(pathInfo.isParentDir('/', '/abc'));
    assert.isTrue(pathInfo.isParentDir('/a', '/a/bc'));
    assert.isTrue(pathInfo.isParentDir('/a', '/a/bc/'));
    assert.isTrue(pathInfo.isParentDir('/a/', '/a/bc'));
    assert.isTrue(pathInfo.isParentDir('/ab/bc', '/ab/bc/cd'));
    assert.isFalse(pathInfo.isParentDir('/a', '/a'));
    assert.isFalse(pathInfo.isParentDir('/ab', '/a'));
    assert.isFalse(pathInfo.isParentDir('/ab', '/abc/d'));
    assert.isFalse(pathInfo.isParentDir('/ab/', '/abc/d'));
    assert.isFalse(pathInfo.isParentDir('/', '/a/b'));
    assert.isFalse(pathInfo.isParentDir('/ab', '/ab/bc/cd'));
    assert.isFalse(pathInfo.isParentDir('/ab/', '/ab/bc/cd'));
  });

  test('getDirname', () => {
    expect(pathInfo.getDirname('')).to.equal('');
    expect(pathInfo.getDirname('/')).to.equal('');
    expect(pathInfo.getDirname('/a.html')).to.equal('');
    expect(pathInfo.getDirname('/a/b.html')).to.equal('/a');
    expect(pathInfo.getDirname('/a/c/b.html')).to.equal('/a/c');
    expect(pathInfo.getDirname('/a/')).to.equal('/a');
    expect(pathInfo.getDirname('/a/b/c')).to.equal('/a/b');
    expect(pathInfo.getDirname('/a/c/b/')).to.equal('/a/c/b');
  });

  test('splitPathIntoLinkedParts', () => {
    expect(pathInfo.splitPathIntoLinkedParts('/')).to.deep.equal([]);

    expect(pathInfo.splitPathIntoLinkedParts('/a/b.html')).to.deep.equal([
      {name: 'a', path: '/a'},
      {name: 'b.html', path: '/a/b.html'},
    ]);

    expect(pathInfo.splitPathIntoLinkedParts('/webgpu/cts.html?q=cts:command_buffer/basic:'))
      .to.deep.equal([
        {name: 'webgpu', path: '/webgpu'},
        {name: 'cts.html?q=cts:command_buffer/basic:', path: '/webgpu/cts.html%3Fq%3Dcts%3Acommand_buffer%2Fbasic%3A'},
      ]);

    expect(pathInfo.splitPathIntoLinkedParts('/a/b.html?x=1#y=2'))
      .to.deep.equal([
        {name: 'a', path: '/a'},
        {name: 'b.html?x=1#y=2', path: '/a/b.html%3Fx%3D1%23y%3D2'},
      ]);

    expect(pathInfo.splitPathIntoLinkedParts('/a.html#x=1'))
      .to.deep.equal([
        {name: 'a.html#x=1', path: '/a.html%23x%3D1'},
      ]);
  });
});

suite('PathPart', () => {
  let sandbox, ppdir, ppfile, ppdirprefixed, ppfileprefixed;

  setup(() => {
    ppdir = fixtureSync(html`<path-part path="/a/b" is-dir="true" navigate="navigate"></path-part>`);
    ppfile = fixtureSync(html`<path-part path="/a/b/c.html" is-dir="false" navigate="navigate"></path-part>`);
    ppdirprefixed = fixtureSync(html`<path-part path="/a/b" is-dir="true" prefix="/foo" navigate="navigate"></path-part>`);
    ppfileprefixed = fixtureSync(html`<path-part path="/a/b/c.html" is-dir="false" prefix="/foo" navigate="navigate"></path-part>`);

    sandbox = sinon.createSandbox();
  });

  teardown(() => {
    sandbox.restore();
  });

  test('is a registered custom element', () => {
    assert.isTrue(new PathPart() instanceof HTMLElement);
    assert.isTrue(document.createElement('path-part') instanceof PathPart);
  });

  suite('static get is()', () => {
    test('path-part', () => {
      assert.equal(PathPart.is, 'path-part');
    });
  });

  suite('PathPart.prototype.*', () => {
    suite('href: computeHref(prefix, path)', () => {
      test('computeHref()', () => {
        assert.equal(typeof PathPart.prototype.computeHref, 'function');
        assert.equal(ppdir.computeHref('x', '/y').pathname, '/x/y');
        assert.equal(ppdir.computeHref('', '/y').pathname, '/y');
        assert.equal(ppdir.computeHref('x', '').pathname, '/x/');
        assert.equal(ppdir.computeHref(undefined, '').pathname, '/');
      });
      test('computeHref([default], path)', () => {
        assert.equal(ppdir.href.pathname, '/a/b');
        assert.equal(ppfile.href.pathname, '/a/b/c.html');
      });
      test('computeHref([foo], path)', () => {
        assert.equal(ppdirprefixed.href.pathname, '/foo/a/b');
        assert.equal(ppfileprefixed.href.pathname, '/foo/a/b/c.html');
      });
      test('computeHref(...) recomputes when prefix changed', () => {
        assert.equal(ppdir.href.pathname, '/a/b');
        assert.equal(ppfile.href.pathname, '/a/b/c.html');

        ppdir.prefix = '/bar';
        ppfile.prefix = '/bar';

        assert.equal(ppdir.href.pathname, '/bar/a/b');
        assert.equal(ppfile.href.pathname, '/bar/a/b/c.html');
      });
    });
    suite('relativePath: computeDisplayableRelativePath(path)', () => {
      // computeDisplayableRelativePath() strips the current page's path, so
      // run these tests from a page path that is not a prefix of the paths
      // under test (as the web-component-tester page path was).
      let originalUrl;
      suiteSetup(() => {
        originalUrl = window.location.href;
        window.history.replaceState(null, '', `/components/test/path.html${window.location.search}`);
      });
      suiteTeardown(() => {
        window.history.replaceState(null, '', originalUrl);
      });

      test('computeDisplayableRelativePath()', () => {
        assert.equal(typeof PathPart.prototype.computeDisplayableRelativePath, 'function');
        assert.equal(ppdir.computeDisplayableRelativePath(''), '/');
        assert.equal(ppdir.computeDisplayableRelativePath('foo'), 'foo/');
      });

      test('computeDisplayableRelativePath(...)', () => {
        assert.equal(ppdir.relativePath, '/a/b/');
      });
      test('computeDisplayableRelativePath(...) (has prefix, is removed)', () => {
        assert.equal(ppdirprefixed.relativePath, '/a/b/');
      });
    });
    suite('styleClass: computePathClass(isDir)', () => {
      test('computePathClass()', () => {
        assert.equal(typeof PathPart.prototype.computePathClass, 'function');
        assert.equal(ppfile.computePathClass(true), 'dir-path');
        assert.equal(ppfile.computePathClass(false), 'file-path');
      });
      test('computePathClass(true)', () => {
        assert.equal(ppdir.styleClass, 'dir-path');
      });
    });
  });
});

// The same tests for PathInfo on a Polymer element and LitPathInfo on a Lit
// element, so that both behave the same. Each host shows two of the computed
// properties in its shadow DOM.
class PolymerPathInfoHost extends PathInfo(PolymerElement) {
  static get template() {
    return polymerHtml`<span id="out">[[encodedPath]] [[pathIsATestFile]]</span>`;
  }
}
window.customElements.define('polymer-path-info-host', PolymerPathInfoHost);

class LitPathInfoHost extends LitPathInfo(LitElement) {
  render() {
    return html`<span id="out">${this.encodedPath} ${this.pathIsATestFile}</span>`;
  }
}
window.customElements.define('lit-path-info-host', LitPathInfoHost);

const pathInfoHosts = [
  {
    name: 'PathInfo (Polymer host)',
    create: () => fixture(html`<polymer-path-info-host></polymer-path-info-host>`),
    createWithPath: () => fixture(html`<polymer-path-info-host path="/a/b.html"></polymer-path-info-host>`),
    rendered: async() => {},
  },
  {
    name: 'LitPathInfo (Lit host)',
    create: () => fixture(html`<lit-path-info-host></lit-path-info-host>`),
    createWithPath: () => fixture(html`<lit-path-info-host path="/a/b.html"></lit-path-info-host>`),
    rendered: el => el.updateComplete,
  },
];

const COMPUTED_PATH_PROPERTIES = ['encodedPath', 'scheme', 'pathIsATestFile', 'pathIsASubfolder', 'pathIsRootDir'];

function computedPathProperties(el, names = COMPUTED_PATH_PROPERTIES) {
  return Object.fromEntries(names.map(name => [name, el[name]]));
}

for (const host of pathInfoHosts) {
  suite(host.name, () => {
    let el;

    function out() {
      return el.shadowRoot.querySelector('#out').textContent;
    }

    setup(async() => {
      el = await host.create();
    });

    test('leaves the computed properties undefined until path is set', () => {
      assert.isUndefined(el.path);
      assert.deepEqual(computedPathProperties(el), {
        encodedPath: undefined,
        scheme: undefined,
        pathIsATestFile: undefined,
        pathIsASubfolder: undefined,
        pathIsRootDir: undefined,
      });
    });

    test('reads path from the path attribute', async() => {
      el = await host.createWithPath();
      assert.strictEqual(el.path, '/a/b.html');
      assert.strictEqual(el.pathIsATestFile, true);
      assert.equal(out(), '/a/b.html true');
    });

    test('computes the properties from path at once, and the view after rendering', async() => {
      el.path = '/a/b/?c';
      assert.strictEqual(el.encodedPath, '/a/b/%3Fc');
      await host.rendered(el);
      assert.equal(out(), '/a/b/%3Fc false');
      el.path = '/a/b.html';
      assert.strictEqual(el.pathIsATestFile, true);
      await host.rendered(el);
      assert.equal(out(), '/a/b.html true');
    });

    test('computes each property from path', () => {
      const cases = [
        ['/', {encodedPath: '/', scheme: 'http', pathIsATestFile: false, pathIsASubfolder: false, pathIsRootDir: true}],
        ['/a/b.html', {encodedPath: '/a/b.html', scheme: 'http', pathIsATestFile: true, pathIsASubfolder: false, pathIsRootDir: false}],
        ['/a/b/', {encodedPath: '/a/b/', scheme: 'http', pathIsATestFile: false, pathIsASubfolder: true, pathIsRootDir: false}],
        ['/a/b.https.html', {encodedPath: '/a/b.https.html', scheme: 'https', pathIsATestFile: true, pathIsASubfolder: false, pathIsRootDir: false}],
        ['/a/b.serviceworker.html?x=1', {encodedPath: '/a/b.serviceworker.html%3Fx%3D1', scheme: 'https', pathIsATestFile: true, pathIsASubfolder: false, pathIsRootDir: false}],
        ['/a/b', {encodedPath: '/a/b', scheme: 'http', pathIsATestFile: false, pathIsASubfolder: true, pathIsRootDir: false}],
      ];
      for (const [path, expected] of cases) {
        el.path = path;
        assert.deepEqual(computedPathProperties(el), expected, path);
      }
    });

    test('computes the properties from an empty or null path', () => {
      // (pathIsATestFile is left out: it depends on the test page's URL.)
      const names = ['encodedPath', 'scheme', 'pathIsASubfolder', 'pathIsRootDir'];
      el.path = '';
      assert.deepEqual(computedPathProperties(el, names),
        {encodedPath: '/', scheme: 'http', pathIsASubfolder: false, pathIsRootDir: ''});
      el.path = null;
      assert.deepEqual(computedPathProperties(el, names),
        {encodedPath: '/', scheme: 'http', pathIsASubfolder: false, pathIsRootDir: null});
    });

    test('has the PathInfo methods', () => {
      assert.isTrue(el.isParentDir('/a', '/a/b'));
      assert.equal(el.getDirname('/a/b.html'), '/a');
      assert.deepEqual(el.splitPathIntoLinkedParts('/a/b.html'), [
        {name: 'a', path: '/a'},
        {name: 'b.html', path: '/a/b.html'},
      ]);
    });
  });
}
