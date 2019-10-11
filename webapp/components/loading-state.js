/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
LoadingState is a behaviour component for indicating when information is
still being loaded (generally, fetched).
*/

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
        notify: true,
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

  retry(f, shouldRetry, num, wait) {
    let count = 0;
    const retry = () => {
      count++;
      return f().catch(err => {
        if (count >= num || !shouldRetry(err)) {
          throw err;
        }
        return new Promise((resolve, reject) => window.setTimeout(
          () => retry().then(resolve, reject),
          wait
        ));
      });
    };
    return retry();
  }
};

export { LoadingState };
