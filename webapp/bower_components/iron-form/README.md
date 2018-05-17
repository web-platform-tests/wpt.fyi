
<!---

This README is automatically generated from the comments in these files:
iron-form.html

Edit those files, and our readme bot will duplicate them over here!
Edit this file, and the bot will squash your changes :)

The bot does some handling of markdown. Please file a bug if it does the wrong
thing! https://github.com/PolymerLabs/tedium/issues

-->

[![Build status](https://travis-ci.org/PolymerElements/iron-form.svg?branch=master)](https://travis-ci.org/PolymerElements/iron-form)
[![Published on webcomponents.org](https://img.shields.io/badge/webcomponents.org-published-blue.svg)](https://www.webcomponents.org/element/PolymerElements/iron-form)

_[Demo and API docs](https://elements.polymer-project.org/elements/iron-form)_


## &lt;iron-form&gt;
`<iron-form>` is a wrapper around the HTML `<form>` element, that can
validate and submit both custom and native HTML elements.

It has two modes: if `allow-redirect` is true, then after the form submission you
will be redirected to the server response. Otherwise, if it is false, it will
use an `iron-ajax` element to submit the form contents to the server.

  Example:

```html
    <iron-form>
      <form method="get" action="/form/handler">
        <input type="text" name="name" value="Batman">
        <input type="checkbox" name="donuts" checked> I like donuts<br>
        <paper-checkbox name="cheese" value="yes" checked></paper-checkbox>
      </form>
    </iron-form>
```

By default, a native `<button>` element (or `input type="submit"`) will submit this form. However, if you
want to submit it from a custom element's click handler, you need to explicitly
call the `iron-form`'s `submit` method.

  Example:

```html
    <paper-button raised onclick="submitForm()">Submit</paper-button>

    function submitForm() {
      document.getElementById('iron-form').submit();
    }
```

### Changes in 2.0
- since type-extensions are not available in 2.0, `<iron-form>` is now a wrapper
around a native `<form>`
- related, since elements are now distributed to the `iron-form`, they no longer
need to implement `IronFormElementBehavior` to register for submission. However
they are required to have a `name` and a `value` attribute (which the behaviour
also added), and to optionally implement the `validate()` method to control
validation of their shadowRoot validatable elements.
- the `serialize` method has been renamed to `serializeForm` (because Polymer 2.0
  is already using a `serialize` method, and we can't stomp over it).
- in `iron-form` 2.x, the `reset` and `submit` methods now accept an `event` as
input, which will be prevented if it exists.
- the `disableNativeValidationUi` property has been removed: because `iron-form`
is no longer a type extension, it can't actually trigger any native UI, so
this property is essentially always true.
- the `contentType` property has been removed in favor of the native [`<form enctype>` 
attribute](https://developer.mozilla.org/en-US/docs/Web/API/HTMLFormElement/enctype);
you can still use the `application/json` value e.g.
```html
<iron-form>
  <form enctype="application/json"> ... </form>
</iron-form>
```
