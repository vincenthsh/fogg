# Changelog

## [3.0.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v2.0.0...cdktf-fogg-helpers-v3.0.0) (2026-06-30)


### ⚠ BREAKING CHANGES

* generated cdktf/terraconstruct components and modules now use cdktn instead of cdktf. Consumers must bump pinned `terraconstructs` cdktf_dependencies to ^0.2.x (helper 2.0.0 peer-requires ^0.2.5); a few components need code fixes for AWS provider v6 namespace changes.

### Features

* Add CDKTF Helpers pkg ([#357](https://github.com/vincenthsh/fogg/issues/357)) ([f1f6347](https://github.com/vincenthsh/fogg/commit/f1f6347490d7094ea9eca726b5073ffc615c59cb))
* CDKTF - Improve cross component remote state access ([#364](https://github.com/vincenthsh/fogg/issues/364)) ([89e4cfc](https://github.com/vincenthsh/fogg/commit/89e4cfc238f8bbbfea45b47e4acac51cd06071a9))
* CDKTF construct utility features ([#376](https://github.com/vincenthsh/fogg/issues/376)) ([b2b498f](https://github.com/vincenthsh/fogg/commit/b2b498fa2df7c61d26e80388da34075442edd086))
* Fogg CDKTF Remote State helpers ([#388](https://github.com/vincenthsh/fogg/issues/388)) ([0efefd1](https://github.com/vincenthsh/fogg/commit/0efefd1be0b9a6b5b7d316eabe5d7f93d5f601d5))
* migrate cdktf templates to cdktn (CDK Terrain) ([#480](https://github.com/vincenthsh/fogg/issues/480)) ([847adb6](https://github.com/vincenthsh/fogg/commit/847adb6ff7c20d60ba8b4e13efab324ba901fbae))
* Update fogg cdktf templates to use helpers pkg ([#360](https://github.com/vincenthsh/fogg/issues/360)) ([3ff113e](https://github.com/vincenthsh/fogg/commit/3ff113e5cc2f67c29300bf39fb6a8d24832bf3cd))
* Use NPMJS instead of GH Pkg npm registry ([#393](https://github.com/vincenthsh/fogg/issues/393)) ([5d546ad](https://github.com/vincenthsh/fogg/commit/5d546ad5e685660a98f846425cd26ddf5417c540))


### Misc

* Bump vitest for security issue ([#382](https://github.com/vincenthsh/fogg/issues/382)) ([acada48](https://github.com/vincenthsh/fogg/commit/acada48fc24d2121eaf13d848b4d8ece2874d166))
* release feat-multi-module-components ([#351](https://github.com/vincenthsh/fogg/issues/351)) ([acec165](https://github.com/vincenthsh/fogg/commit/acec165cb6e2f32fa31678da7d6fa5f47aac5f7e))
* release feat-multi-module-components ([#361](https://github.com/vincenthsh/fogg/issues/361)) ([85824a2](https://github.com/vincenthsh/fogg/commit/85824a25f278fd20d7366b7471dbdbfeee6bbeb9))
* release feat-multi-module-components ([#365](https://github.com/vincenthsh/fogg/issues/365)) ([03ec444](https://github.com/vincenthsh/fogg/commit/03ec4441e81be0af3af637918214e04b12b6b180))
* release feat-multi-module-components ([#377](https://github.com/vincenthsh/fogg/issues/377)) ([b07241d](https://github.com/vincenthsh/fogg/commit/b07241de924e25f98e81cce53ce41184d79df1d4))
* release feat-multi-module-components ([#381](https://github.com/vincenthsh/fogg/issues/381)) ([bb6556f](https://github.com/vincenthsh/fogg/commit/bb6556fbef461488d2691b7aa68e31399fef4865))
* release feat-multi-module-components ([#384](https://github.com/vincenthsh/fogg/issues/384)) ([1671c98](https://github.com/vincenthsh/fogg/commit/1671c986593266ae077e793093cab95369ccdc12))
* release feat-multi-module-components ([#390](https://github.com/vincenthsh/fogg/issues/390)) ([82be07e](https://github.com/vincenthsh/fogg/commit/82be07e3e84d8ff69c4512414f46566e968e3322))
* release feat-multi-module-components ([#392](https://github.com/vincenthsh/fogg/issues/392)) ([848c35b](https://github.com/vincenthsh/fogg/commit/848c35bc16e2ea43e8eaf1f0749e218d562b2743))
* release feat-multi-module-components ([#394](https://github.com/vincenthsh/fogg/issues/394)) ([93a1aee](https://github.com/vincenthsh/fogg/commit/93a1aee3ff8d0887b5c5365dd6380a664f035e4a))
* release feat-multi-module-components ([#397](https://github.com/vincenthsh/fogg/issues/397)) ([cc7bedd](https://github.com/vincenthsh/fogg/commit/cc7bedd72c951ff43704a5e703844caeae1c8f8d))
* Update fogg-types ([#383](https://github.com/vincenthsh/fogg/issues/383)) ([830063a](https://github.com/vincenthsh/fogg/commit/830063a98c04fea88021c33f1bc2d884ba3b5983))


### BugFixes

* missing publishConfig for scoped packages ([#396](https://github.com/vincenthsh/fogg/issues/396)) ([b799469](https://github.com/vincenthsh/fogg/commit/b7994695a533f10471880650edf89771d2a5d05d))
* missing registry-url for publish action ([#395](https://github.com/vincenthsh/fogg/issues/395)) ([2934f14](https://github.com/vincenthsh/fogg/commit/2934f148dada55750a54f5a4ba5b6c38ea90fc30))
* remotestateproxy output wrappers bug ([#389](https://github.com/vincenthsh/fogg/issues/389)) ([4a9a6a5](https://github.com/vincenthsh/fogg/commit/4a9a6a52b7b2094a32cc9aa6621ba4468d448b89))

## [1.5.2](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.5.1...cdktf-fogg-helpers-v1.5.2) (2025-02-20)


### BugFixes

* missing publishConfig for scoped packages ([#396](https://github.com/vincenthsh/fogg/issues/396)) ([b799469](https://github.com/vincenthsh/fogg/commit/b7994695a533f10471880650edf89771d2a5d05d))

## [1.5.1](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.5.0...cdktf-fogg-helpers-v1.5.1) (2025-02-20)


### BugFixes

* missing registry-url for publish action ([#395](https://github.com/vincenthsh/fogg/issues/395)) ([2934f14](https://github.com/vincenthsh/fogg/commit/2934f148dada55750a54f5a4ba5b6c38ea90fc30))

## [1.5.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.4.1...cdktf-fogg-helpers-v1.5.0) (2025-02-20)


### Features

* Use NPMJS instead of GH Pkg npm registry ([#393](https://github.com/vincenthsh/fogg/issues/393)) ([5d546ad](https://github.com/vincenthsh/fogg/commit/5d546ad5e685660a98f846425cd26ddf5417c540))

## [1.4.1](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.4.0...cdktf-fogg-helpers-v1.4.1) (2025-02-15)


### BugFixes

* remotestateproxy output wrappers bug ([#389](https://github.com/vincenthsh/fogg/issues/389)) ([4a9a6a5](https://github.com/vincenthsh/fogg/commit/4a9a6a52b7b2094a32cc9aa6621ba4468d448b89))

## [1.4.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.3.1...cdktf-fogg-helpers-v1.4.0) (2025-02-14)


### Features

* Fogg CDKTF Remote State helpers ([#388](https://github.com/vincenthsh/fogg/issues/388)) ([0efefd1](https://github.com/vincenthsh/fogg/commit/0efefd1be0b9a6b5b7d316eabe5d7f93d5f601d5))


### Misc

* Update fogg-types ([#383](https://github.com/vincenthsh/fogg/issues/383)) ([830063a](https://github.com/vincenthsh/fogg/commit/830063a98c04fea88021c33f1bc2d884ba3b5983))

## [1.3.1](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.3.0...cdktf-fogg-helpers-v1.3.1) (2025-02-05)


### Misc

* Bump vitest for security issue ([#382](https://github.com/vincenthsh/fogg/issues/382)) ([acada48](https://github.com/vincenthsh/fogg/commit/acada48fc24d2121eaf13d848b4d8ece2874d166))

## [1.3.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.2.0...cdktf-fogg-helpers-v1.3.0) (2025-02-01)


### Features

* CDKTF construct utility features ([#376](https://github.com/vincenthsh/fogg/issues/376)) ([b2b498f](https://github.com/vincenthsh/fogg/commit/b2b498fa2df7c61d26e80388da34075442edd086))

## [1.2.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.1.0...cdktf-fogg-helpers-v1.2.0) (2025-01-05)


### Features

* CDKTF - Improve cross component remote state access ([#364](https://github.com/vincenthsh/fogg/issues/364)) ([89e4cfc](https://github.com/vincenthsh/fogg/commit/89e4cfc238f8bbbfea45b47e4acac51cd06071a9))

## [1.1.0](https://github.com/vincenthsh/fogg/compare/cdktf-fogg-helpers-v1.0.0...cdktf-fogg-helpers-v1.1.0) (2025-01-01)


### Features

* Update fogg cdktf templates to use helpers pkg ([#360](https://github.com/vincenthsh/fogg/issues/360)) ([3ff113e](https://github.com/vincenthsh/fogg/commit/3ff113e5cc2f67c29300bf39fb6a8d24832bf3cd))

## 1.0.0 (2024-12-31)


### Features

* Add CDKTF Helpers pkg ([#357](https://github.com/vincenthsh/fogg/issues/357)) ([f1f6347](https://github.com/vincenthsh/fogg/commit/f1f6347490d7094ea9eca726b5073ffc615c59cb))
