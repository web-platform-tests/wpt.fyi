/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<loading-state>` is a behaviour component for indicating when information is
still being loaded (generally, fetched).
*/
const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="loading-state">

</dom-module>`;

document.head.appendChild($_documentContainer.content);
// eslint-disable-next-line no-unused-vars
const LoadingState = (superClass) => class extends superClass {
  static get properties() {
    return {
      loadingCount: {
        type: Number,
        value: 0,
        observer: 'loadingCountChanged',
      },
      isLoading: {
        type: Boolean,
        computed: 'computeIsLoading(loadingCount)',
      },
      onLoadingComplete: Function,
    };
  }

  computeIsLoading(loadingCount) {
    return !!loadingCount;
  }

  loadingCountChanged(now, then) {
    if (now === 0 && then > 0 && this.onLoadingComplete) {
      this.onLoadingComplete();
    }
  }

  async load(promise, opt_errHandler) {
    this.loadingCount++;
    try {
      return await promise;
    } catch (e) {
      // eslint-disable-next-line no-console
      console.log(`Failed to load: ${e}`);
      if (opt_errHandler) {
        opt_errHandler(e);
      }
    } finally {
      this.loadingCount--;
    }
  }
};

export { LoadingState };