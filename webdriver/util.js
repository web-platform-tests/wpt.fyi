const puppeteer = require('puppeteer');

/**
 * Pierces shadow DOM in the page for the given selector.
 * @param {puppeteer.Page} page
 * @param {puppeteer.ElementHandle} root Root element to query. If null, page is
 *    used.
 * @param {Array<string>} selectors Selectors in each level of shadow DOM.
 * @returns {?puppeteer.ElementHandle}
 */
const shadowQuery = async function(page, root, ...selectors) {
  if (!selectors || !selectors.length) {
    return;
  }
  const [first, ...others] = selectors;
  let result = await (root || page).$(first);
  for (const next of others) {
    result = await page.evaluateHandle(
      (e, selector) => e.shadowRoot.querySelector(selector),
      result,
      next);
  }
  return result;
}
exports.shadowQuery = shadowQuery;

/**
 * Pierces shadow DOM in the given element for the given selector.
 * @param {puppeteer.ElementHandle} element
 * @param {string} selector Selector for the shadow DOM
 * @returns {Array<puppeteer.ElementHandle>}
 */
const shadowQueryAll = async function(page, element, selector) {
  return await page.evaluateHandle(
    (e, selector) => e.shadowRoot.querySelectorAll(selector),
    element,
    selector);
}
exports.shadowQueryAll = shadowQueryAll;


