// +build small

package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const githubCommitsComparisonBody = `{
  "url": "https://api.github.com/repos/web-platform-tests/wpt/compare/4b370b4c0baed4174377ffc0c4b51208ad26c62c...bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
  "html_url": "https://github.com/web-platform-tests/wpt/compare/4b370b4c0baed4174377ffc0c4b51208ad26c62c...bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
  "permalink_url": "https://github.com/web-platform-tests/wpt/compare/web-platform-tests:4b370b4...web-platform-tests:bb9c0ce",
  "diff_url": "https://github.com/web-platform-tests/wpt/compare/4b370b4c0baed4174377ffc0c4b51208ad26c62c...bb9c0ce15e70865a9d01a9f6dbb8568913a15d76.diff",
  "patch_url": "https://github.com/web-platform-tests/wpt/compare/4b370b4c0baed4174377ffc0c4b51208ad26c62c...bb9c0ce15e70865a9d01a9f6dbb8568913a15d76.patch",
  "base_commit": {
    "sha": "4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "node_id": "MDY6Q29tbWl0MzYxODEzMzo0YjM3MGI0YzBiYWVkNDE3NDM3N2ZmYzBjNGI1MTIwOGFkMjZjNjJj",
    "commit": {
      "author": {
        "name": "autofoolip",
        "email": "40241672+autofoolip@users.noreply.github.com",
        "date": "2018-10-02T11:02:07Z"
      },
      "committer": {
        "name": "Philip Jägenstedt",
        "email": "philip@foolip.org",
        "date": "2018-10-02T11:02:07Z"
      },
      "message": "Update interfaces/payment-request.idl (#13307)\n\nSource: https://github.com/tidoust/reffy-reports/blob/e2599eb/whatwg/idl/payment-request.idl\r\nBuild: https://travis-ci.org/tidoust/reffy-reports/builds/435978193",
      "tree": {
        "sha": "401f8afc4f0c186888bc75e6492b27a956f19f50",
        "url": "https://api.github.com/repos/web-platform-tests/wpt/git/trees/401f8afc4f0c186888bc75e6492b27a956f19f50"
      },
      "url": "https://api.github.com/repos/web-platform-tests/wpt/git/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
      "comment_count": 0,
      "verification": {
        "verified": false,
        "reason": "unsigned",
        "signature": null,
        "payload": null
      }
    },
    "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "html_url": "https://github.com/web-platform-tests/wpt/commit/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "comments_url": "https://api.github.com/repos/web-platform-tests/wpt/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c/comments",
    "author": {
      "login": "autofoolip",
      "id": 40241672,
      "node_id": "MDQ6VXNlcjQwMjQxNjcy",
      "avatar_url": "https://avatars2.githubusercontent.com/u/40241672?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/autofoolip",
      "html_url": "https://github.com/autofoolip",
      "followers_url": "https://api.github.com/users/autofoolip/followers",
      "following_url": "https://api.github.com/users/autofoolip/following{/other_user}",
      "gists_url": "https://api.github.com/users/autofoolip/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/autofoolip/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/autofoolip/subscriptions",
      "organizations_url": "https://api.github.com/users/autofoolip/orgs",
      "repos_url": "https://api.github.com/users/autofoolip/repos",
      "events_url": "https://api.github.com/users/autofoolip/events{/privacy}",
      "received_events_url": "https://api.github.com/users/autofoolip/received_events",
      "type": "User",
      "site_admin": false
    },
    "committer": {
      "login": "foolip",
      "id": 498917,
      "node_id": "MDQ6VXNlcjQ5ODkxNw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/498917?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/foolip",
      "html_url": "https://github.com/foolip",
      "followers_url": "https://api.github.com/users/foolip/followers",
      "following_url": "https://api.github.com/users/foolip/following{/other_user}",
      "gists_url": "https://api.github.com/users/foolip/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/foolip/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/foolip/subscriptions",
      "organizations_url": "https://api.github.com/users/foolip/orgs",
      "repos_url": "https://api.github.com/users/foolip/repos",
      "events_url": "https://api.github.com/users/foolip/events{/privacy}",
      "received_events_url": "https://api.github.com/users/foolip/received_events",
      "type": "User",
      "site_admin": false
    },
    "parents": [
      {
        "sha": "07a61a1900c33f39a3e170d612afecbb77417bdb",
        "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/07a61a1900c33f39a3e170d612afecbb77417bdb",
        "html_url": "https://github.com/web-platform-tests/wpt/commit/07a61a1900c33f39a3e170d612afecbb77417bdb"
      }
    ]
  },
  "merge_base_commit": {
    "sha": "4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "node_id": "MDY6Q29tbWl0MzYxODEzMzo0YjM3MGI0YzBiYWVkNDE3NDM3N2ZmYzBjNGI1MTIwOGFkMjZjNjJj",
    "commit": {
      "author": {
        "name": "autofoolip",
        "email": "40241672+autofoolip@users.noreply.github.com",
        "date": "2018-10-02T11:02:07Z"
      },
      "committer": {
        "name": "Philip Jägenstedt",
        "email": "philip@foolip.org",
        "date": "2018-10-02T11:02:07Z"
      },
      "message": "Update interfaces/payment-request.idl (#13307)\n\nSource: https://github.com/tidoust/reffy-reports/blob/e2599eb/whatwg/idl/payment-request.idl\r\nBuild: https://travis-ci.org/tidoust/reffy-reports/builds/435978193",
      "tree": {
        "sha": "401f8afc4f0c186888bc75e6492b27a956f19f50",
        "url": "https://api.github.com/repos/web-platform-tests/wpt/git/trees/401f8afc4f0c186888bc75e6492b27a956f19f50"
      },
      "url": "https://api.github.com/repos/web-platform-tests/wpt/git/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
      "comment_count": 0,
      "verification": {
        "verified": false,
        "reason": "unsigned",
        "signature": null,
        "payload": null
      }
    },
    "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "html_url": "https://github.com/web-platform-tests/wpt/commit/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
    "comments_url": "https://api.github.com/repos/web-platform-tests/wpt/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c/comments",
    "author": {
      "login": "autofoolip",
      "id": 40241672,
      "node_id": "MDQ6VXNlcjQwMjQxNjcy",
      "avatar_url": "https://avatars2.githubusercontent.com/u/40241672?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/autofoolip",
      "html_url": "https://github.com/autofoolip",
      "followers_url": "https://api.github.com/users/autofoolip/followers",
      "following_url": "https://api.github.com/users/autofoolip/following{/other_user}",
      "gists_url": "https://api.github.com/users/autofoolip/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/autofoolip/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/autofoolip/subscriptions",
      "organizations_url": "https://api.github.com/users/autofoolip/orgs",
      "repos_url": "https://api.github.com/users/autofoolip/repos",
      "events_url": "https://api.github.com/users/autofoolip/events{/privacy}",
      "received_events_url": "https://api.github.com/users/autofoolip/received_events",
      "type": "User",
      "site_admin": false
    },
    "committer": {
      "login": "foolip",
      "id": 498917,
      "node_id": "MDQ6VXNlcjQ5ODkxNw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/498917?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/foolip",
      "html_url": "https://github.com/foolip",
      "followers_url": "https://api.github.com/users/foolip/followers",
      "following_url": "https://api.github.com/users/foolip/following{/other_user}",
      "gists_url": "https://api.github.com/users/foolip/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/foolip/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/foolip/subscriptions",
      "organizations_url": "https://api.github.com/users/foolip/orgs",
      "repos_url": "https://api.github.com/users/foolip/repos",
      "events_url": "https://api.github.com/users/foolip/events{/privacy}",
      "received_events_url": "https://api.github.com/users/foolip/received_events",
      "type": "User",
      "site_admin": false
    },
    "parents": [
      {
        "sha": "07a61a1900c33f39a3e170d612afecbb77417bdb",
        "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/07a61a1900c33f39a3e170d612afecbb77417bdb",
        "html_url": "https://github.com/web-platform-tests/wpt/commit/07a61a1900c33f39a3e170d612afecbb77417bdb"
      }
    ]
  },
  "status": "ahead",
  "ahead_by": 2,
  "behind_by": 0,
  "total_commits": 2,
  "commits": [
    {
      "sha": "63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
      "node_id": "MDY6Q29tbWl0MzYxODEzMzo2MzY0MWZiNWZlOTk0MmVmZDNlYjgxYTVjODMxYWI3ZjI0N2U0MGMy",
      "commit": {
        "author": {
          "name": "Maja Kabus",
          "email": "kabusm@google.com",
          "date": "2018-10-01T20:36:55Z"
        },
        "committer": {
          "name": "Philip Jägenstedt",
          "email": "philip@foolip.org",
          "date": "2018-10-02T11:30:33Z"
        },
        "message": "isXXX methods added to TrustedTypePolicyFactory\n\nisHTML(), isScript(), isScriptURL() and isURL() added to\nTrustedTypePolicyFactory class as part of Trusted Types API update to\ncurrent JS polyfill.\nThe methods require additional code to be fully matched to polyfill\nexpected behaviour.\n\nAdded a helper private method GetWrapperTypeInfoFromScriptValue.\n\nBug: 739170\nChange-Id: I027e43ab6432405c686255a4d0ce24248c59a4dc\nReviewed-on: https://chromium-review.googlesource.com/1238433\nCommit-Queue: Daniel Vogelheim <vogelheim@chromium.org>\nReviewed-by: Daniel Vogelheim <vogelheim@chromium.org>\nCr-Commit-Position: refs/heads/master@{#595527}",
        "tree": {
          "sha": "cb25ffc1d961027eb6c69ed0657bb2f92d75dd6b",
          "url": "https://api.github.com/repos/web-platform-tests/wpt/git/trees/cb25ffc1d961027eb6c69ed0657bb2f92d75dd6b"
        },
        "url": "https://api.github.com/repos/web-platform-tests/wpt/git/commits/63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
        "comment_count": 0,
        "verification": {
          "verified": false,
          "reason": "unsigned",
          "signature": null,
          "payload": null
        }
      },
      "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
      "html_url": "https://github.com/web-platform-tests/wpt/commit/63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
      "comments_url": "https://api.github.com/repos/web-platform-tests/wpt/commits/63641fb5fe9942efd3eb81a5c831ab7f247e40c2/comments",
      "author": null,
      "committer": {
        "login": "foolip",
        "id": 498917,
        "node_id": "MDQ6VXNlcjQ5ODkxNw==",
        "avatar_url": "https://avatars1.githubusercontent.com/u/498917?v=4",
        "gravatar_id": "",
        "url": "https://api.github.com/users/foolip",
        "html_url": "https://github.com/foolip",
        "followers_url": "https://api.github.com/users/foolip/followers",
        "following_url": "https://api.github.com/users/foolip/following{/other_user}",
        "gists_url": "https://api.github.com/users/foolip/gists{/gist_id}",
        "starred_url": "https://api.github.com/users/foolip/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/foolip/subscriptions",
        "organizations_url": "https://api.github.com/users/foolip/orgs",
        "repos_url": "https://api.github.com/users/foolip/repos",
        "events_url": "https://api.github.com/users/foolip/events{/privacy}",
        "received_events_url": "https://api.github.com/users/foolip/received_events",
        "type": "User",
        "site_admin": false
      },
      "parents": [
        {
          "sha": "4b370b4c0baed4174377ffc0c4b51208ad26c62c",
          "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/4b370b4c0baed4174377ffc0c4b51208ad26c62c",
          "html_url": "https://github.com/web-platform-tests/wpt/commit/4b370b4c0baed4174377ffc0c4b51208ad26c62c"
        }
      ]
    },
    {
      "sha": "bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "node_id": "MDY6Q29tbWl0MzYxODEzMzpiYjljMGNlMTVlNzA4NjVhOWQwMWE5ZjZkYmI4NTY4OTEzYTE1ZDc2",
      "commit": {
        "author": {
          "name": "Josh Matthews",
          "email": "josh@joshmatthews.net",
          "date": "2018-10-02T11:03:00Z"
        },
        "committer": {
          "name": "jgraham",
          "email": "james@hoppipolla.co.uk",
          "date": "2018-10-02T12:05:40Z"
        },
        "message": "Rename test file that conflicts with existing css test file.",
        "tree": {
          "sha": "732c94e3e3dac20e75ecea9d61f577e758ed9c1b",
          "url": "https://api.github.com/repos/web-platform-tests/wpt/git/trees/732c94e3e3dac20e75ecea9d61f577e758ed9c1b"
        },
        "url": "https://api.github.com/repos/web-platform-tests/wpt/git/commits/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
        "comment_count": 0,
        "verification": {
          "verified": false,
          "reason": "unsigned",
          "signature": null,
          "payload": null
        }
      },
      "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "html_url": "https://github.com/web-platform-tests/wpt/commit/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "comments_url": "https://api.github.com/repos/web-platform-tests/wpt/commits/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/comments",
      "author": {
        "login": "jdm",
        "id": 27658,
        "node_id": "MDQ6VXNlcjI3NjU4",
        "avatar_url": "https://avatars1.githubusercontent.com/u/27658?v=4",
        "gravatar_id": "",
        "url": "https://api.github.com/users/jdm",
        "html_url": "https://github.com/jdm",
        "followers_url": "https://api.github.com/users/jdm/followers",
        "following_url": "https://api.github.com/users/jdm/following{/other_user}",
        "gists_url": "https://api.github.com/users/jdm/gists{/gist_id}",
        "starred_url": "https://api.github.com/users/jdm/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/jdm/subscriptions",
        "organizations_url": "https://api.github.com/users/jdm/orgs",
        "repos_url": "https://api.github.com/users/jdm/repos",
        "events_url": "https://api.github.com/users/jdm/events{/privacy}",
        "received_events_url": "https://api.github.com/users/jdm/received_events",
        "type": "User",
        "site_admin": false
      },
      "committer": {
        "login": "jgraham",
        "id": 294864,
        "node_id": "MDQ6VXNlcjI5NDg2NA==",
        "avatar_url": "https://avatars1.githubusercontent.com/u/294864?v=4",
        "gravatar_id": "",
        "url": "https://api.github.com/users/jgraham",
        "html_url": "https://github.com/jgraham",
        "followers_url": "https://api.github.com/users/jgraham/followers",
        "following_url": "https://api.github.com/users/jgraham/following{/other_user}",
        "gists_url": "https://api.github.com/users/jgraham/gists{/gist_id}",
        "starred_url": "https://api.github.com/users/jgraham/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/jgraham/subscriptions",
        "organizations_url": "https://api.github.com/users/jgraham/orgs",
        "repos_url": "https://api.github.com/users/jgraham/repos",
        "events_url": "https://api.github.com/users/jgraham/events{/privacy}",
        "received_events_url": "https://api.github.com/users/jgraham/received_events",
        "type": "User",
        "site_admin": false
      },
      "parents": [
        {
          "sha": "63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
          "url": "https://api.github.com/repos/web-platform-tests/wpt/commits/63641fb5fe9942efd3eb81a5c831ab7f247e40c2",
          "html_url": "https://github.com/web-platform-tests/wpt/commit/63641fb5fe9942efd3eb81a5c831ab7f247e40c2"
        }
      ]
    }
  ],
  "files": [
    {
      "sha": "b82abea82a603f9a523fd8ee0a4d61cd0b7349c6",
      "filename": "css/css-shapes/shape-outside/values/shape-outside-inset-0010.html",
      "status": "renamed",
      "additions": 0,
      "deletions": 0,
      "changes": 0,
      "blob_url": "https://github.com/web-platform-tests/wpt/blob/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/css/css-shapes/shape-outside/values/shape-outside-inset-0010.html",
      "raw_url": "https://github.com/web-platform-tests/wpt/raw/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/css/css-shapes/shape-outside/values/shape-outside-inset-0010.html",
      "contents_url": "https://api.github.com/repos/web-platform-tests/wpt/contents/css/css-shapes/shape-outside/values/shape-outside-inset-0010.html?ref=bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "previous_filename": "css/css-shapes/shape-outside/values/shape-outside-inset-010.html"
    },
    {
      "sha": "a162d84cd820051d6c5868c35b58cd347b0026e5",
      "filename": "trusted-types/TrustedTypePolicyFactory-createPolicy-createXYZTests.tentative.html",
      "status": "modified",
      "additions": 16,
      "deletions": 8,
      "changes": 24,
      "blob_url": "https://github.com/web-platform-tests/wpt/blob/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/trusted-types/TrustedTypePolicyFactory-createPolicy-createXYZTests.tentative.html",
      "raw_url": "https://github.com/web-platform-tests/wpt/raw/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/trusted-types/TrustedTypePolicyFactory-createPolicy-createXYZTests.tentative.html",
      "contents_url": "https://api.github.com/repos/web-platform-tests/wpt/contents/trusted-types/TrustedTypePolicyFactory-createPolicy-createXYZTests.tentative.html?ref=bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "patch": "@@ -7,8 +7,10 @@\n   //HTML tests\n   function createHTMLTest(policyName, policy, expectedHTML, t) {\n     let p = window.TrustedTypes.createPolicy(policyName, policy);\n-assert_true(p.createHTML('whatever') instanceof TrustedHTML);\n-    assert_equals(p.createHTML('whatever') + \"\", expectedHTML);\n+    let html = p.createHTML('whatever');\n+    assert_true(html instanceof TrustedHTML);\n+    assert_true(TrustedTypes.isHTML(html));\n+    assert_equals(html + \"\", expectedHTML);\n   }\n \n   test(t => {\n@@ -77,8 +79,10 @@\n   //Script tests\n function createScriptTest(policyName, policy, expectedScript, t) {\n     let p = window.TrustedTypes.createPolicy(policyName, policy);\n-    assert_true(p.createScript('whatever') instanceof TrustedScript);\n-    assert_equals(p.createScript('whatever') + \"\", expectedScript);\n+    let script = p.createScript('whatever');\n+    assert_true(script instanceof TrustedScript);\n+   assert_true(TrustedTypes.isScript(script));\n+    assert_equals(script + \"\", expectedScript);\n   }\n \n   test(t => {\n@@ -150,8 +154,10 @@\n   //ScriptURL tests\n   function createScriptURLTest(policyName, policy, expectedScriptURL, t) {\n     let p = window.TrustedTypes.createPolicy(policyName, policy);\n-    assert_true(p.createScriptURL(INPUTS.SCRIPTURL) instanceof TrustedScriptURL);\n-    assert_equals(p.createScriptURL(INPUTS.SCRIPTURL) + \"\", expectedScriptURL);\n+    let scriptUrl = p.createScriptURL(INPUTS.SCRIPTURL);\n+    assert_true(scriptUrl instanceof TrustedScriptURL);\n+    assert_true(TrustedTypes.isScriptURL(scriptUrl));\n+    assert_equals(scriptUrl + \"\", expectedScriptURL);\n   }\n \n   test(t => {\n@@ -223,8 +229,10 @@\n   //URL tests\n   function createURLTest(policyName, policy, expectedURL, t) {\n     let p = window.TrustedTypes.createPolicy(policyName, policy);\n-    assert_true(p.createURL(INPUTS.URL) instanceof TrustedURL);\n-    assert_equals(p.createURL(INPUTS.URL) + \"\", expectedURL);\n+    let url = p.createURL(INPUTS.URL);\n+    assert_true(url instanceof TrustedURL);\n+    assert_true(TrustedTypes.isURL(url));\n+    assert_equals(url + \"\", expectedURL);\n   }\n \n   test(t => {"
    },
    {
      "sha": "9b48fa7fede81b5d2e2c79d9fc115b56c759cb00",
      "filename": "trusted-types/TrustedTypePolicyFactory-isXXX.tentative.html",
      "status": "added",
      "additions": 118,
      "deletions": 0,
      "changes": 118,
      "blob_url": "https://github.com/web-platform-tests/wpt/blob/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/trusted-types/TrustedTypePolicyFactory-isXXX.tentative.html",
      "raw_url": "https://github.com/web-platform-tests/wpt/raw/bb9c0ce15e70865a9d01a9f6dbb8568913a15d76/trusted-types/TrustedTypePolicyFactory-isXXX.tentative.html",
      "contents_url": "https://api.github.com/repos/web-platform-tests/wpt/contents/trusted-types/TrustedTypePolicyFactory-isXXX.tentative.html?ref=bb9c0ce15e70865a9d01a9f6dbb8568913a15d76",
      "patch": "@@ -0,0 +1,118 @@\n+<!DOCTYPE html>\n+<script src=\"/resources/testharness.js\"></script>\n+<script src=\"/resources/testharnessreport.js\"></script>\n+<script src=\"support/helper.sub.js\"></script>\n+\n+<meta http-equiv=\"Content-Security-Policy\" content=\"trusted-types\">\n+<body>\n+<script>\n+  // Policy settings for all tests\n+  const noopPolicy = {\n+    'createHTML': (s) => s,\n+    'createScriptURL': (s) => s,\n+    'createURL': (s) => s,\n+    'createScript': (s) => s,\n+  };\n+\n+  // isHTML tests\n+  test(t => {\n+    const p = TrustedTypes.createPolicy('html', noopPolicy);\n+    let html = p.createHTML(INPUTS.HTML);\n+\n+    assert_true(TrustedTypes.isHTML(html));\n+    let html2 = Object.create(html);\n+\n+    // instanceof can pass, but we rely on isHTML\n+    assert_true(html2 instanceof TrustedHTML);\n+    assert_false(TrustedTypes.isHTML(html2));\n+\n+    let html3 = Object.assign({}, html, {toString: () => 'fake'});\n+\n+    assert_false(TrustedTypes.isHTML(html3));\n+  }, 'TrustedTypePolicyFactory.isHTML requires the object to be created via policy.');\n+\n+  // isScript tests\n+  test(t => {\n+    const p = TrustedTypes.createPolicy('script', noopPolicy);\n+    let script = p.createScript(INPUTS.SCRIPT);\n+\n+    assert_true(TrustedTypes.isScript(script));\n+    let script2 = Object.create(script);\n+\n+    // instanceof can pass, but we rely on isScript\n+    assert_true(script2 instanceof TrustedScript);\n+    assert_false(TrustedTypes.isScript(script2));\n+\n+ let script3 = Object.assign({}, script, {toString: () => 'fake'});\n+\n+    assert_false(TrustedTypes.isScript(script3));\n+  }, 'TrustedTypePolicyFactory.isScript requires the object to becreated via policy.');\n+\n+  // isScriptURL tests\n+  test(t => {\n+    const p = TrustedTypes.createPolicy('script_url', noopPolicy);\n+    let script = p.createScriptURL(INPUTS.SCRIPTURL);\n+\n+    assert_true(TrustedTypes.isScriptURL(script));\n+    let script2 = Object.create(script);\n+\n+    // instanceof can pass, but we rely on isScript\n+    assert_true(script2 instanceof TrustedScriptURL);\n+    assert_false(TrustedTypes.isScriptURL(script2));\n+\n+    let script3 = Object.assign({}, script, {toString: () => 'fake'});\n+\n+    assert_false(TrustedTypes.isScriptURL(script3));\n+  }, 'TrustedTypePolicyFactory.isScriptURL requires the object to be created via policy.');\n+\n+  // isURL tests\n+  test(t => {\n+    const p = TrustedTypes.createPolicy('url', noopPolicy);\n+    let url = p.createURL(INPUTS.URL);\n+\n+    assert_true(TrustedTypes.isURL(url));\n+    let url2 = Object.create(url);\n+\n+    // instanceof can pass, but we rely on isScript\n+    assert_true(url2 instanceof TrustedURL);\n+    assert_false(TrustedTypes.isURL(url2));\n+\n+    let url3 = Object.assign({}, url, {toString: () => 'fake'});\n+\n+    assert_false(TrustedTypes.isURL(url3));\n+  }, 'TrustedTypePolicyFactory.isURL requires the object to be created via policy.');\n+\n+  // Redefinition tests\n+  // TODO(vogelheim): Implement TrustedTypes (& policy objects) as 'frozen'.\n+/*  test(t => {\n+    assert_throws(new TypeError(), _ => {\n+      TrustedTypes.isHTML = () => true;\n+    });\n+\n+    assert_false(TrustedTypes.isHTML({}));\n+  }, 'TrustedTypePolicyFactory.IsHTML cannot be redefined.');\n+\n+  test(t => {\n+    assert_throws(new TypeError(), _ => {\n+      TrustedTypes.isScript = () => true;\n+    });\n+\n+    assert_false(TrustedTypes.isScript({}));\n+  }, 'TrustedTypePolicyFactory.isScript cannot be redefined.');\n+\n+  test(t => {\n+    assert_throws(new TypeError(), _ => {\n+      TrustedTypes.isScriptURL = () => true;\n+    });\n+\n+    assert_false(TrustedTypes.isScriptURL({}));\n+  }, 'TrustedTypePolicyFactory.isScriptURL cannot be redefined.');\n+\n+  test(t => {\n+  assert_throws(new TypeError(), _ => {\n+      TrustedTypes.isURL = () => true;\n+    });\n+\n+    assert_false(TrustedTypes.isURL({}));\n+  }, 'TrustedTypePolicyFactory.isURL cannot be redefined.');*/\n+</script>"
    }
  ]
}
`

func TestGetRenames_ResponseParsesCorrectly(t *testing.T) {
	comparison := new(githubCommitsComparison)
	err := json.Unmarshal([]byte(githubCommitsComparisonBody), comparison)
	assert.Nil(t, err)
	assert.Len(t, comparison.Files, 3)
	assert.Equal(t, "renamed", *comparison.Files[0].Status)
}
