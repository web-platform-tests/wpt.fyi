/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/**
 * LoadingState is a behaviour component for indicating when information is
 * still being loaded (generally, fetched).
 */
const LoadingState = (superClass) => class extends LoadingStateMethods(superClass) {
  static get properties() {
    return {
      /**
       * The number of active loading operations.
       */
      loadingCount: {
        type: Number,
        value: 0,
        observer: '_loadingCountChanged',
      },
      /**
       * Whether the component is currently loading data.
       * Computed based on `loadingCount`.
       */
      isLoading: {
        type: Boolean,
        value: false,
        computed: '_computeIsLoading(loadingCount)',
        readOnly: true,
      },
      /**
       * A callback function to be executed when loading is complete.
       */
      onLoadingComplete: {
        type: Function,
      },
    };
  }
};

/**
 * The logic shared by LoadingState (Polymer) and LitLoadingState (Lit).
 */
const LoadingStateMethods = (superClass) => class extends superClass {
  /**
   * Computes the `isLoading` property based on `loadingCount`.
   * @param {number} loadingCount The current loading count.
   * @return {boolean} True if loading, false otherwise.
   */
  _computeIsLoading(loadingCount) {
    return loadingCount > 0;
  }

  /**
   * Observer for `loadingCount` changes.
   * Calls `onLoadingComplete` when loading finishes.
   * @param {number} now The new loading count.
   * @param {number} then The previous loading count.
   */
  _loadingCountChanged(now, then) {
    if (now === 0 && then > 0 && this.onLoadingComplete) {
      this.onLoadingComplete();
    }
  }

  /**
   * Tracks a promise, incrementing `loadingCount` while it's pending.
   * @param {Promise} promise The promise to track.
   * @param {Function} opt_errHandler An optional error handler.
   * @return {Promise} A promise that resolves/rejects with the original promise.
   */
  async load(promise, opt_errHandler) {
    this.loadingCount++;
    try {
      return await promise;
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error(`Failed to load: ${e}`);
      if (opt_errHandler) {
        opt_errHandler(e);
      }
    } finally {
      this.loadingCount--;
    }
  }

  /**
   * Retries a function with exponential backoff.
   * @param {Function} f The function to retry.
   * @param {Function} shouldRetry A function that determines if retrying should continue.
   * @param {number} num The maximum number of retries.
   * @param {number} wait The initial wait time in milliseconds.
   * @return {Promise} A promise that resolves with the result of `f` or rejects.
   */
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

/**
 * LitLoadingState is LoadingState for Lit elements.
 *
 * As with the Polymer property observer, `onLoadingComplete` is called
 * synchronously, from the `loadingCount` setter, when the count goes from a
 * positive number back to 0. `isLoading` is read-only and derived from
 * `loadingCount`.
 */
const LitLoadingState = (superClass) => class extends LoadingStateMethods(superClass) {
  static properties = {
    /**
     * The number of active loading operations.
     */
    loadingCount: {
      type: Number,
      attribute: 'loading-count',
      noAccessor: true,
    },
    /**
     * A callback function to be executed when loading is complete.
     */
    onLoadingComplete: {
      attribute: false,
    },
  };

  #loadingCount;

  constructor() {
    super();
    this.loadingCount = 0;
  }

  get loadingCount() {
    return this.#loadingCount;
  }

  set loadingCount(now) {
    const then = this.#loadingCount;
    // The same change check as Polymer's (NaN doesn't change to NaN).
    if (now === then || (now !== now && then !== then)) {
      return;
    }
    this.#loadingCount = now;
    this.requestUpdate('loadingCount', then);
    this._loadingCountChanged(now, then);
  }

  /**
   * Whether the component is currently loading data.
   */
  get isLoading() {
    return this._computeIsLoading(this.loadingCount);
  }
};

export { LitLoadingState, LoadingState };
