# Changelog

All notable changes to this project will be documented in this file.

## [0.5.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.5.0...v0.5.1) (2026-07-23)

### 🔧 Miscellaneous

* **adl:** refresh agent.yaml defaults from ADL CLI v0.54.0 ([#106](https://github.com/inference-gateway/google-calendar-agent/issues/106)) ([53abd4d](https://github.com/inference-gateway/google-calendar-agent/commit/53abd4d86ee6c614c483d30091dd24985a204f0c))
* **deps:** bump ADL CLI to v0.54.0 ([#107](https://github.com/inference-gateway/google-calendar-agent/issues/107)) ([2245a4f](https://github.com/inference-gateway/google-calendar-agent/commit/2245a4f87182fe11beec4c31340caedf53c6e0d8))

### 🔨 Miscellaneous

* **deps:** bump the github-actions group with 3 updates ([#105](https://github.com/inference-gateway/google-calendar-agent/issues/105)) ([9bd7420](https://github.com/inference-gateway/google-calendar-agent/commit/9bd74200e669f2dd0c116ed60d61c26010b3c510))
* **deps:** bump the gomod group with 3 updates ([#104](https://github.com/inference-gateway/google-calendar-agent/issues/104)) ([3a59377](https://github.com/inference-gateway/google-calendar-agent/commit/3a5937764fee2e63ea29b3d8eb8944cd53001deb))

## [0.5.0](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.29...v0.5.0) (2026-07-17)

### ✨ Features

* **telemetry:** add OpenTelemetry support via agent.yaml manifest ([#101](https://github.com/inference-gateway/google-calendar-agent/issues/101)) ([f5e4f45](https://github.com/inference-gateway/google-calendar-agent/commit/f5e4f45b80490dd685531ab40e2742477ad3bc8a)), references [#103](https://github.com/inference-gateway/google-calendar-agent/issues/103)

### 🐛 Bug Fixes

* wrong char in manifest ([aea1894](https://github.com/inference-gateway/google-calendar-agent/commit/aea1894e76d3ae8b0f1e8b3811301ca28aa98a66))

### 👷 CI

* **claude:** change effort to max ([d7490f6](https://github.com/inference-gateway/google-calendar-agent/commit/d7490f632fadfb4eae8ca758b6852b178eab15d3))
* **claude:** remove system prompt - use default community maintained prompt ([586f621](https://github.com/inference-gateway/google-calendar-agent/commit/586f62172712f1d11f705b94d2546050b547252d))
* **claude:** standardize workflow + task-based branch prefix ([325c342](https://github.com/inference-gateway/google-calendar-agent/commit/325c3421b036f272654d01e080fc78764e8c7dff))
* **deps:** downgrade task version from 3.51.1 to 3.48.0 in workflows and manifest ([4764c34](https://github.com/inference-gateway/google-calendar-agent/commit/4764c34df25a9d728efec2755fafccbb3bf673a7))
* **release:** update semantic release and plugins to latest versions with local installation ([18dcc0f](https://github.com/inference-gateway/google-calendar-agent/commit/18dcc0f3977720267a57d3b7f56f7aebf54389cd))

### 📚 Documentation

* author spec.examples and spec.documentation in agent.yaml ([#91](https://github.com/inference-gateway/google-calendar-agent/issues/91)) ([5a25dfb](https://github.com/inference-gateway/google-calendar-agent/commit/5a25dfb9e58ae278016726ba0dfcf9a84e66665a)), closes [#90](https://github.com/inference-gateway/google-calendar-agent/issues/90)

### 🔧 Miscellaneous

* **adl:** Refresh agent.yaml defaults from ADL CLI v0.50.2 ([#92](https://github.com/inference-gateway/google-calendar-agent/issues/92)) ([57f94a9](https://github.com/inference-gateway/google-calendar-agent/commit/57f94a933f9f6c2c14d205d8c5a1a60a2ed2a5bd))
* **deps:** bump ADL CLI to v0.47.1 ([#80](https://github.com/inference-gateway/google-calendar-agent/issues/80)) ([026b31e](https://github.com/inference-gateway/google-calendar-agent/commit/026b31e21ec910e1dfd8aac40f6d06e5765cba33))
* **deps:** bump ADL CLI to v0.52.0 ([#102](https://github.com/inference-gateway/google-calendar-agent/issues/102)) ([b707f8d](https://github.com/inference-gateway/google-calendar-agent/commit/b707f8d44e50adf9d4837b1e2140cb674c0c0c04))
* **deps:** bump ADL CLI v0.39.3 -> v0.40.0 ([#57](https://github.com/inference-gateway/google-calendar-agent/issues/57)) ([4db779e](https://github.com/inference-gateway/google-calendar-agent/commit/4db779e6c91463478f7060f0734051b539f5a524))
* **deps:** bump ADL CLI v0.40.0 -> v0.43.2 ([#60](https://github.com/inference-gateway/google-calendar-agent/issues/60)) ([2059402](https://github.com/inference-gateway/google-calendar-agent/commit/205940232494c5e8f3c4fe1f62bac037e3379e86))
* **deps:** bump ADL CLI v0.43.2 -> v0.44.0 ([#66](https://github.com/inference-gateway/google-calendar-agent/issues/66)) ([2091dee](https://github.com/inference-gateway/google-calendar-agent/commit/2091dee7adcca3d9c203f0e5ece6384b730a8fa1))
* **deps:** bump ADL CLI v0.44.0 -> v0.46.0 ([#74](https://github.com/inference-gateway/google-calendar-agent/issues/74)) ([94ed59e](https://github.com/inference-gateway/google-calendar-agent/commit/94ed59e6ab6811408dd1b48bc569b05c9d19b4ab))
* **deps:** bump ADL CLI v0.46.0 -> v0.46.5 ([#77](https://github.com/inference-gateway/google-calendar-agent/issues/77)) ([0967547](https://github.com/inference-gateway/google-calendar-agent/commit/096754721cf4833c732247e2d9690e4e8d341f70))
* **deps:** bump ADL CLI v0.46.5 -> v0.47.0 ([#79](https://github.com/inference-gateway/google-calendar-agent/issues/79)) ([b995993](https://github.com/inference-gateway/google-calendar-agent/commit/b995993e7d670bae74765d3a77f71dcf4e1375d6))
* **deps:** bump ADL CLI v0.47.1 -> v0.48.0 ([#81](https://github.com/inference-gateway/google-calendar-agent/issues/81)) ([a4ff173](https://github.com/inference-gateway/google-calendar-agent/commit/a4ff173e2d647d21659e27b51efbf170f526b827))
* **deps:** bump ADL CLI v0.48.0 -> v0.48.1 ([#83](https://github.com/inference-gateway/google-calendar-agent/issues/83)) ([b44829f](https://github.com/inference-gateway/google-calendar-agent/commit/b44829f09a6303265865186de3aab671c3a6e33c))
* **deps:** bump ADL CLI v0.48.1 -> v0.48.4 ([#85](https://github.com/inference-gateway/google-calendar-agent/issues/85)) ([01953f4](https://github.com/inference-gateway/google-calendar-agent/commit/01953f4134d8f1802243ab6bd3d19aff9b8dc153))
* **deps:** bump ADL CLI v0.48.4 -> v0.48.5 ([#87](https://github.com/inference-gateway/google-calendar-agent/issues/87)) ([f734378](https://github.com/inference-gateway/google-calendar-agent/commit/f7343785d17cefa346629bbc28ec9e0e7364e957))
* **deps:** bump ADL CLI v0.48.5 -> v0.49.0 ([#88](https://github.com/inference-gateway/google-calendar-agent/issues/88)) ([739d40b](https://github.com/inference-gateway/google-calendar-agent/commit/739d40b58453389cb04fee9cf21367bd47289beb))
* **deps:** bump ADL CLI v0.49.0 -> v0.50.2 ([#93](https://github.com/inference-gateway/google-calendar-agent/issues/93)) ([7b1e6f0](https://github.com/inference-gateway/google-calendar-agent/commit/7b1e6f0439f50890218bcc44a71f67f9ea301fbb))
* **deps:** bump ADL CLI v0.50.2 -> v0.51.0 ([#95](https://github.com/inference-gateway/google-calendar-agent/issues/95)) ([0edebc6](https://github.com/inference-gateway/google-calendar-agent/commit/0edebc6aeb3cb1ab7af37cc2061605e21be16e34))
* **deps:** bump ADL CLI v0.51.0 -> v0.51.4 ([#99](https://github.com/inference-gateway/google-calendar-agent/issues/99)) ([f2e15ae](https://github.com/inference-gateway/google-calendar-agent/commit/f2e15aeac914d172d77d3afcc1fd927433857901))
* **deps:** bump docker/setup-qemu-action version v4.0.0 -> v4.1.0 ([874cc47](https://github.com/inference-gateway/google-calendar-agent/commit/874cc4764c878af6e434899ebf7ddd6dcd207517))
* **flox:** add missing lock file changes ([9ad634f](https://github.com/inference-gateway/google-calendar-agent/commit/9ad634f92e1541ce82074f41277ab1780557a94b))
* **flox:** downgrade deps ([5b9fc40](https://github.com/inference-gateway/google-calendar-agent/commit/5b9fc40afa163ace61c56a70a07103140007b895))
* **schema:** update adl schema to latest ([bdf2285](https://github.com/inference-gateway/google-calendar-agent/commit/bdf2285f45f6b28e709e71e6df9505d43b756d1c))

### 🔨 Miscellaneous

* **deps:** bump actions/checkout in the github-actions group ([#67](https://github.com/inference-gateway/google-calendar-agent/issues/67)) ([c0fa8ef](https://github.com/inference-gateway/google-calendar-agent/commit/c0fa8ef632201b27037801fdce8edcc9b79e446b))
* **deps:** bump actions/setup-go in the github-actions group ([#98](https://github.com/inference-gateway/google-calendar-agent/issues/98)) ([9b2f2f7](https://github.com/inference-gateway/google-calendar-agent/commit/9b2f2f7b856f83e16f3ee85eea337f383821ee78))
* **deps:** bump anthropics/claude-code-action ([#61](https://github.com/inference-gateway/google-calendar-agent/issues/61)) ([d83ffd9](https://github.com/inference-gateway/google-calendar-agent/commit/d83ffd9accb2eba6469e7da1ffa66d33eda82d77))
* **deps:** bump anthropics/claude-code-action ([#65](https://github.com/inference-gateway/google-calendar-agent/issues/65)) ([f55f7e4](https://github.com/inference-gateway/google-calendar-agent/commit/f55f7e488d08cb3bda64e96f2db3ec9ced6f3c3c))
* **deps:** bump anthropics/claude-code-action ([#82](https://github.com/inference-gateway/google-calendar-agent/issues/82)) ([6215d44](https://github.com/inference-gateway/google-calendar-agent/commit/6215d443483d806dbb7543315669470ae7deb8a3))
* **deps:** bump anthropics/claude-code-action ([#84](https://github.com/inference-gateway/google-calendar-agent/issues/84)) ([0bf0353](https://github.com/inference-gateway/google-calendar-agent/commit/0bf0353f6fa0e0f5ec9e2eca789f3e368d76789c))
* **deps:** bump anthropics/claude-code-action ([#96](https://github.com/inference-gateway/google-calendar-agent/issues/96)) ([e554f8e](https://github.com/inference-gateway/google-calendar-agent/commit/e554f8eeaef044da7ca6b62f68706bf5d216b607))
* **deps:** bump github.com/inference-gateway/adk in the gomod group ([#86](https://github.com/inference-gateway/google-calendar-agent/issues/86)) ([230adaa](https://github.com/inference-gateway/google-calendar-agent/commit/230adaafd7f2dfc3129472f895fda47ebbafba87))
* **deps:** bump github.com/inference-gateway/adk in the gomod group ([#97](https://github.com/inference-gateway/google-calendar-agent/issues/97)) ([2bef340](https://github.com/inference-gateway/google-calendar-agent/commit/2bef340031256872c8d6da5657c2dfb86c045d5c))
* **deps:** bump github.com/quic-go/quic-go from 0.59.0 to 0.59.1 ([#63](https://github.com/inference-gateway/google-calendar-agent/issues/63)) ([01b956a](https://github.com/inference-gateway/google-calendar-agent/commit/01b956ac07b086d20629d7fc3bd4f740af4c1e84))
* **deps:** bump github.com/sethvargo/go-envconfig in the gomod group ([#103](https://github.com/inference-gateway/google-calendar-agent/issues/103)) ([a757bce](https://github.com/inference-gateway/google-calendar-agent/commit/a757bce5b7189c0d3b73cf997fa56c45f97c5fc6))
* **deps:** bump Go version from 1.26.2 to 1.26.4 in agent configuration ([553b607](https://github.com/inference-gateway/google-calendar-agent/commit/553b60794995f379f7dfc58bde13ad64784f3377))
* **deps:** bump google.golang.org/api in the gomod group ([#56](https://github.com/inference-gateway/google-calendar-agent/issues/56)) ([9a6cd4c](https://github.com/inference-gateway/google-calendar-agent/commit/9a6cd4c994c880d9930d015b4e2fec29eb7956f7))
* **deps:** bump google.golang.org/api in the gomod group ([#59](https://github.com/inference-gateway/google-calendar-agent/issues/59)) ([c78a040](https://github.com/inference-gateway/google-calendar-agent/commit/c78a04029298e2c525f9b0ec6cdc880052058ddd))
* **deps:** bump google.golang.org/api in the gomod group ([#62](https://github.com/inference-gateway/google-calendar-agent/issues/62)) ([1534ff4](https://github.com/inference-gateway/google-calendar-agent/commit/1534ff45d3eadaac7e1613ba25d6f352a469c795))
* **deps:** bump google.golang.org/api in the gomod group ([#64](https://github.com/inference-gateway/google-calendar-agent/issues/64)) ([40b3658](https://github.com/inference-gateway/google-calendar-agent/commit/40b3658754d531d8e074f935370c16ef7f3d9b38))
* **deps:** bump google.golang.org/api in the gomod group ([#68](https://github.com/inference-gateway/google-calendar-agent/issues/68)) ([cc58e14](https://github.com/inference-gateway/google-calendar-agent/commit/cc58e141f1694b892eea10fbc15589fc4698ea3d))
* **deps:** bump google.golang.org/api in the gomod group ([#70](https://github.com/inference-gateway/google-calendar-agent/issues/70)) ([479d1b5](https://github.com/inference-gateway/google-calendar-agent/commit/479d1b5c7901521a987357f30372c7d1fd3b81e6))
* **deps:** bump inference-gateway/infer-action ([#100](https://github.com/inference-gateway/google-calendar-agent/issues/100)) ([a7750e4](https://github.com/inference-gateway/google-calendar-agent/commit/a7750e429012440c3ad0ed3f0789c36fa3467aa8))
* **deps:** bump inference-gateway/infer-action ([#89](https://github.com/inference-gateway/google-calendar-agent/issues/89)) ([1f69362](https://github.com/inference-gateway/google-calendar-agent/commit/1f693629aa24a8138eee6ec9fb2d22cec23e4102))
* **deps:** bump inference-gateway/infer-action ([#94](https://github.com/inference-gateway/google-calendar-agent/issues/94)) ([496061a](https://github.com/inference-gateway/google-calendar-agent/commit/496061a8ea2bf65e5579e33b5c51ccc53cf956c8))
* **deps:** bump the github-actions group with 2 updates ([#58](https://github.com/inference-gateway/google-calendar-agent/issues/58)) ([3ac7589](https://github.com/inference-gateway/google-calendar-agent/commit/3ac758902406d2d1138610570dba97536299db7a))
* **deps:** bump the github-actions group with 2 updates ([#69](https://github.com/inference-gateway/google-calendar-agent/issues/69)) ([79e18a7](https://github.com/inference-gateway/google-calendar-agent/commit/79e18a7e4ae2e2661c1fd12f7de0de162b61e9b0))
* **deps:** bump the github-actions group with 3 updates ([#72](https://github.com/inference-gateway/google-calendar-agent/issues/72)) ([cee6207](https://github.com/inference-gateway/google-calendar-agent/commit/cee6207a8960dd21a980cc8fccd2a03dc98281df))
* **deps:** bump the github-actions group with 4 updates ([#71](https://github.com/inference-gateway/google-calendar-agent/issues/71)) ([c04e1d1](https://github.com/inference-gateway/google-calendar-agent/commit/c04e1d106e95e72bb9d465070f563b4334f14eba))
* **deps:** bump the gomod group with 2 updates ([#73](https://github.com/inference-gateway/google-calendar-agent/issues/73)) ([b0335af](https://github.com/inference-gateway/google-calendar-agent/commit/b0335afd71df912d3b94e74bb1dec234782e78dd))

## [0.4.29](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.28...v0.4.29) (2026-05-26)

### 🔧 Miscellaneous

* **deps:** Bump ADL CLI v0.39.2 -> v0.39.3 ([#54](https://github.com/inference-gateway/google-calendar-agent/issues/54)) ([e606d5e](https://github.com/inference-gateway/google-calendar-agent/commit/e606d5e2ad0898d130309ec91631a7a35576fa74))
* Replace em dash with regular dash ([6116175](https://github.com/inference-gateway/google-calendar-agent/commit/6116175e2185e806625ecfefa844c9e95be95ee2))

## [0.4.28](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.27...v0.4.28) (2026-05-24)

### 🔧 Miscellaneous

* **deps:** Bump ADL CLI v0.39.1 -> v0.39.2 ([#53](https://github.com/inference-gateway/google-calendar-agent/issues/53)) ([0002cd1](https://github.com/inference-gateway/google-calendar-agent/commit/0002cd133a0965ff87c839ae10e5554359877734))

## [0.4.27](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.26...v0.4.27) (2026-05-24)

### ♻️ Improvements

* **Taskfile:** Remove explicit flags from generate command ([8120e89](https://github.com/inference-gateway/google-calendar-agent/commit/8120e892ef27006db8a99130c2ef880dc19c5adc))
* **tools:** Panic-safe arg parsing, all-day events, test coverage ([#46](https://github.com/inference-gateway/google-calendar-agent/issues/46)) ([47eb895](https://github.com/inference-gateway/google-calendar-agent/commit/47eb895f6416ff6828bf1c9babe2d676dbed1802))

### 👷 CI

* **claude:** Simplify conditions for triggering Claude Code actions ([5cc64d1](https://github.com/inference-gateway/google-calendar-agent/commit/5cc64d1bad16da29cf0e78247828be6b95ab17bb))
* **deps:** Update claude-code-action to version 1.0.130 ([19052cf](https://github.com/inference-gateway/google-calendar-agent/commit/19052cfb8ab81a90fc8d79ea90179d98c412bab6))

### 🔧 Miscellaneous

* **adl:** Refresh agent.yaml defaults from ADL CLI v0.33.1 ([#30](https://github.com/inference-gateway/google-calendar-agent/issues/30)) ([7e8f4d1](https://github.com/inference-gateway/google-calendar-agent/commit/7e8f4d1cd89cb55ec793cd38f51be052c2d4a697))
* **adl:** Refresh agent.yaml defaults from ADL CLI v0.36.0 ([#41](https://github.com/inference-gateway/google-calendar-agent/issues/41)) ([b020b6e](https://github.com/inference-gateway/google-calendar-agent/commit/b020b6ee9d06b8990e212c210390a03df0f15c9c))
* **adl:** Refresh agent.yaml defaults from ADL CLI v0.38.1 ([#49](https://github.com/inference-gateway/google-calendar-agent/issues/49)) ([23c7947](https://github.com/inference-gateway/google-calendar-agent/commit/23c7947de8cfc57192633fcaf955d60fb057d09d))
* **dependabot:** Update golang and ubuntu version ignore rules in dependabot configuration ([02d33b8](https://github.com/inference-gateway/google-calendar-agent/commit/02d33b8b5ed816473ad903de6bbe3ade58e90aae))
* **deps:** Add ignore rule for golang dependency in dependabot configuration ([a5d590e](https://github.com/inference-gateway/google-calendar-agent/commit/a5d590eb733e94a8fd1417495a0548df0d61b959))
* **deps:** Bump ADL CLI v0.30.10 -> v0.31.0 ([#29](https://github.com/inference-gateway/google-calendar-agent/issues/29)) ([a1c9557](https://github.com/inference-gateway/google-calendar-agent/commit/a1c95574c86ac5ed6d12c653e5ed917dc49da370))
* **deps:** Bump ADL CLI v0.31.0 -> v0.34.0 ([#32](https://github.com/inference-gateway/google-calendar-agent/issues/32)) ([4b3488f](https://github.com/inference-gateway/google-calendar-agent/commit/4b3488f93ed972e2ce9b7b91625a516186070b2e))
* **deps:** Bump ADL CLI v0.34.0 -> v0.34.1 ([#36](https://github.com/inference-gateway/google-calendar-agent/issues/36)) ([4f3205d](https://github.com/inference-gateway/google-calendar-agent/commit/4f3205d3994519443a5a99d89d8c27e6916378c2))
* **deps:** Bump ADL CLI v0.34.1 -> v0.34.2 ([#39](https://github.com/inference-gateway/google-calendar-agent/issues/39)) ([92290bb](https://github.com/inference-gateway/google-calendar-agent/commit/92290bb494980849004fde63937145f5e57f1495))
* **deps:** Bump ADL CLI v0.34.2 -> v0.36.1 ([#42](https://github.com/inference-gateway/google-calendar-agent/issues/42)) ([649e18c](https://github.com/inference-gateway/google-calendar-agent/commit/649e18c5c673294e3f10c8bf42b398ede6daf6f2))
* **deps:** Bump ADL CLI v0.36.1 -> v0.36.2 ([#45](https://github.com/inference-gateway/google-calendar-agent/issues/45)) ([2a61be6](https://github.com/inference-gateway/google-calendar-agent/commit/2a61be6a9459d57c24689d297995852dd76cf38e))
* **deps:** Bump ADL CLI v0.36.2 -> v0.36.4 ([#47](https://github.com/inference-gateway/google-calendar-agent/issues/47)) ([67bc044](https://github.com/inference-gateway/google-calendar-agent/commit/67bc0446fc7ea02e95bbf244f421dbafcf793ab6))
* **deps:** Bump ADL CLI v0.36.4 -> v0.38.1 ([#48](https://github.com/inference-gateway/google-calendar-agent/issues/48)) ([56076f3](https://github.com/inference-gateway/google-calendar-agent/commit/56076f3f3e91eeac2e2fbb899fa2c4e0f830ad06))
* **deps:** Bump ADL CLI v0.38.1 -> v0.39.0 ([#51](https://github.com/inference-gateway/google-calendar-agent/issues/51)) ([c856a60](https://github.com/inference-gateway/google-calendar-agent/commit/c856a6061faceb91b3086fcf81966c25dd754d83))
* **deps:** Bump ADL CLI v0.39.0 -> v0.39.1 ([#52](https://github.com/inference-gateway/google-calendar-agent/issues/52)) ([88d92dd](https://github.com/inference-gateway/google-calendar-agent/commit/88d92ddede43a71700266ad9fd92f4ba192c8c5f))
* **flox:** Generate manifest lock file ([3e8902c](https://github.com/inference-gateway/google-calendar-agent/commit/3e8902c9024110d4ee7e64b1afe9d2ab050b21af))
* **license:** Update license to Apache 2.0 ([3f2aa30](https://github.com/inference-gateway/google-calendar-agent/commit/3f2aa303fcde9d972f612ed6c2ec88e4009a7680))
* Update .adl-ignore ([0cf6060](https://github.com/inference-gateway/google-calendar-agent/commit/0cf60603ed12835f25d765f29a66369fb1a8f198))

### 🔨 Miscellaneous

* **deps:** Bump anthropics/claude-code-action ([#26](https://github.com/inference-gateway/google-calendar-agent/issues/26)) ([3b3d18f](https://github.com/inference-gateway/google-calendar-agent/commit/3b3d18f064415a5c51695b0f99f9e221ccde6350))
* **deps:** Bump anthropics/claude-code-action ([#28](https://github.com/inference-gateway/google-calendar-agent/issues/28)) ([0e2e4dc](https://github.com/inference-gateway/google-calendar-agent/commit/0e2e4dca9c301023fc576b9a2bb3c599450675e2))
* **deps:** Bump github.com/inference-gateway/adk in the gomod group ([#27](https://github.com/inference-gateway/google-calendar-agent/issues/27)) ([0075e2a](https://github.com/inference-gateway/google-calendar-agent/commit/0075e2a2bfd12d7a07112d7bd7fe5146d7e31f7d))
* **deps:** Bump github.com/inference-gateway/adk in the gomod group ([#37](https://github.com/inference-gateway/google-calendar-agent/issues/37)) ([474a6c8](https://github.com/inference-gateway/google-calendar-agent/commit/474a6c83d4bb5457729e5110baf0ce41a456a4b9))
* **deps:** Bump the github-actions group with 2 updates ([#40](https://github.com/inference-gateway/google-calendar-agent/issues/40)) ([e7d226a](https://github.com/inference-gateway/google-calendar-agent/commit/e7d226aa606eb27fa1d4a8fe1fff8de90da1f6e5))

## [0.4.26](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.25...v0.4.26) (2026-05-20)

### 👷 CI

* **deps:** Update installation steps for golangci-lint and task in CI/CD workflows ([e03d156](https://github.com/inference-gateway/google-calendar-agent/commit/e03d1564340ed54bf8dc2a9dcbac750247f61ae9))

### 🔧 Miscellaneous

* **deps:** Bump ADL CLI to version 0.30.10 ([67e13a5](https://github.com/inference-gateway/google-calendar-agent/commit/67e13a58bf5cae005355936f016689322eb13ba6))

## [0.4.25](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.24...v0.4.25) (2026-05-19)

### ♻️ Improvements

* Migrate to latest ADL schema (CLI 0.30.7) ([#24](https://github.com/inference-gateway/google-calendar-agent/issues/24)) ([eeebcc4](https://github.com/inference-gateway/google-calendar-agent/commit/eeebcc4c11db37640a30b2eecd4f7d9eb604ee01))

### 👷 CI

* **dependabot:** Add dependabot to help with dependecies upgrades ([763fc1f](https://github.com/inference-gateway/google-calendar-agent/commit/763fc1ff2d8c5c668d72b8f660d34034ea20064d))
* **deps:** Bump the github-actions group with 5 updates ([#22](https://github.com/inference-gateway/google-calendar-agent/issues/22)) ([6da8fff](https://github.com/inference-gateway/google-calendar-agent/commit/6da8fff14ca8f8b1c901b33a31f870c57cea901d))
* Enable display report for Claude Code action ([4233df6](https://github.com/inference-gateway/google-calendar-agent/commit/4233df64d4b3a7a6fbafbabd877e054a6f82c952))
* Update create-github-app-token action to v3.2.0 ([c7a8264](https://github.com/inference-gateway/google-calendar-agent/commit/c7a82649e25744bbce3c4af0ad3e7afbc7bec05d))

### 🔧 Miscellaneous

* Create CODEOWNERS ([f724e2f](https://github.com/inference-gateway/google-calendar-agent/commit/f724e2ff302c0ce3dcb6cef43de6b57e5c5adcff))
* Remove outdated issue templates for bug reports, feature requests, and refactor requests ([3dcbdcc](https://github.com/inference-gateway/google-calendar-agent/commit/3dcbdcc35e2bdd78e734e31f43bf05ef580fbd48))

### 🔨 Miscellaneous

* **deps:** Bump golang in the docker group ([#21](https://github.com/inference-gateway/google-calendar-agent/issues/21)) ([080c1fd](https://github.com/inference-gateway/google-calendar-agent/commit/080c1fd4a44e85f82208cece972bd89a1517068c))
* **deps:** Bump the gomod group with 3 updates ([#23](https://github.com/inference-gateway/google-calendar-agent/issues/23)) ([676401c](https://github.com/inference-gateway/google-calendar-agent/commit/676401cf60f8d5597ba81f85fa480af92e07026a))

## [0.4.24](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.23...v0.4.24) (2026-05-07)

### ♻️ Improvements

* Rename all instances of deepseek-chat to deepseek-v4-flash ([840f47c](https://github.com/inference-gateway/google-calendar-agent/commit/840f47c0d6eb92f11be2e2015a004e0380dde6e2))
* Update task installation method in CI and CD workflows ([d17adb4](https://github.com/inference-gateway/google-calendar-agent/commit/d17adb4ec935657da99afeaff541768c15aa8826))

### 👷 CI

* Bump claude code action ([ce7db09](https://github.com/inference-gateway/google-calendar-agent/commit/ce7db090b86804edb9360e7b74a7bad36a4dec96))
* **deps:** Bump golangci-lint to latest ([0efd5e4](https://github.com/inference-gateway/google-calendar-agent/commit/0efd5e4099603f2f0dcdc615a64efd8204f6fbfb))
* Update golangci-lint installation script to use the latest URL and version ([ffbe1ab](https://github.com/inference-gateway/google-calendar-agent/commit/ffbe1abaa2fd9e6f4cb859854f6b6cc0734a97bb))

## [0.4.23](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.22...v0.4.23) (2026-04-17)

### 🔧 Miscellaneous

* **deps:** Bump ADL CLI to 0.27.8 ([6826375](https://github.com/inference-gateway/google-calendar-agent/commit/6826375770da0e367140cafc200c18fcd8ea0296))

## [0.4.22](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.21...v0.4.22) (2026-04-17)

### 👷 CI

* **workflow:** Update Claude Code Action configuration and dependencies ([61c8db5](https://github.com/inference-gateway/google-calendar-agent/commit/61c8db5136250fc0c603b61fc1ce828b7e5035db))

### 🔧 Miscellaneous

* Bump devcontainer dependecies ([e201fa6](https://github.com/inference-gateway/google-calendar-agent/commit/e201fa650bd9ed77f6300e9b530692ef9eb0596c))
* **deps:** Bump ADL CLI to 0.27.6 and re-generate the codebase ([d4e7fea](https://github.com/inference-gateway/google-calendar-agent/commit/d4e7feadcb1b517b18c5c08f77f031b92700ceed))

## [0.4.21](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.20...v0.4.21) (2026-01-27)

### 🔧 Miscellaneous

* **deps:** Bump versions to latest ([3de78d3](https://github.com/inference-gateway/google-calendar-agent/commit/3de78d3e93534a7acc5844ff99b5a52abe7db7e6))

## [0.4.20](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.19...v0.4.20) (2025-10-20)

### 🔧 Miscellaneous

* **deps:** Update ADL CLI version and dependencies ([3bff53b](https://github.com/inference-gateway/google-calendar-agent/commit/3bff53b8f1ec3f31b32ed5e2ec01d6aa941b5271))

## [0.4.19](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.18...v0.4.19) (2025-10-07)

### 🔧 Miscellaneous

* **deps:** Update dependencies and regenerate files with ADL CLI v0.23.0 ([2cb12b5](https://github.com/inference-gateway/google-calendar-agent/commit/2cb12b5f4bcfc9b655ea0831c9da27e985edb48e))

## [0.4.18](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.17...v0.4.18) (2025-09-26)

### ♻️ Improvements

* Bump ADK version to 0.11.1 ([7ab7bdb](https://github.com/inference-gateway/google-calendar-agent/commit/7ab7bdb758d9e915a003a44f600bfb24d7b74976))

## [0.4.17](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.16...v0.4.17) (2025-09-26)

### ♻️ Improvements

* Update agent metadata file references and regenerate files for ADL CLI v0.21.6 ([#19](https://github.com/inference-gateway/google-calendar-agent/issues/19)) ([7717b32](https://github.com/inference-gateway/google-calendar-agent/commit/7717b32972f4dd429555210c58e115e0e3937d41))

## [0.4.16](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.15...v0.4.16) (2025-09-22)

### ♻️ Improvements

* Make this agent conform to the ADL CLI for easier maintenance ([#18](https://github.com/inference-gateway/google-calendar-agent/issues/18)) ([acec809](https://github.com/inference-gateway/google-calendar-agent/commit/acec809bce84550601707bb5c9db87de609e6247))

## [0.4.15](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.14...v0.4.15) (2025-08-28)

### ♻️ Improvements

* Integrate LLM client and update environment configuration for A2A agent ([ffb5d88](https://github.com/inference-gateway/google-calendar-agent/commit/ffb5d88fb61a1d5c22ddd4ee7ba1105ee9cd1a8c))

## [0.4.14](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.13...v0.4.14) (2025-08-28)

### 🐛 Bug Fixes

* Add default background task handler to server ([665709b](https://github.com/inference-gateway/google-calendar-agent/commit/665709b4c6c347964b19e902bb29b9f7dce917d5))

## [0.4.13](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.12...v0.4.13) (2025-08-28)

### 🐛 Bug Fixes

* **agent:** Set agent details and add default streaming task handler ([5d5f8f3](https://github.com/inference-gateway/google-calendar-agent/commit/5d5f8f31a3752a02365de63555fa3648f475132d))

## [0.4.12](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.11...v0.4.12) (2025-08-28)

### ♻️ Improvements

* **adk:** Upgrade to latest ADK server version v0.9.0 ([#17](https://github.com/inference-gateway/google-calendar-agent/issues/17)) ([d4418d9](https://github.com/inference-gateway/google-calendar-agent/commit/d4418d9e610f60dcf6006a41e22285e4211c7398))

### 🔧 Miscellaneous

* Remove redundant reminder about using `task check` before pushing changes ([3aa3449](https://github.com/inference-gateway/google-calendar-agent/commit/3aa34492571023dd000d005f21706d3a9cca5a9a))

## [0.4.11](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.10...v0.4.11) (2025-08-28)

### 👷 CI

* Add Claude Code GitHub Workflow ([#15](https://github.com/inference-gateway/google-calendar-agent/issues/15)) ([99659b2](https://github.com/inference-gateway/google-calendar-agent/commit/99659b25e5d2ae5fdaa0571696a73a8e34d2905b))

### 📚 Documentation

* Add CLAUDE.md for project guidance and development instructions ([60c24f5](https://github.com/inference-gateway/google-calendar-agent/commit/60c24f5951d6ceeb1b5ef44e07e10c0d018f45a8))

### 🔧 Miscellaneous

* Add initial configuration files for flox environment ([afa3cff](https://github.com/inference-gateway/google-calendar-agent/commit/afa3cffcfa0a4f91083d892878c601f03fd45a2f))
* Add issue templates for bug reports, feature requests, and refactor requests ([40b1a41](https://github.com/inference-gateway/google-calendar-agent/commit/40b1a4151f83fc21276d35117694b0a68f08f56b))

## [0.4.10](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.9...v0.4.10) (2025-07-21)

### 🐛 Bug Fixes

* Add extra_files section for agent.json in Docker configurations ([b817b16](https://github.com/inference-gateway/google-calendar-agent/commit/b817b16854916e95812865e021d6d662b2d3f5dd))

## [0.4.9](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.8...v0.4.9) (2025-07-21)

### 🐛 Bug Fixes

* Add missing agent.json copy in Dockerfiles ([9d4354d](https://github.com/inference-gateway/google-calendar-agent/commit/9d4354da3301d68e1210c84f018dff1b391bbb2c))

## [0.4.8](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.7...v0.4.8) (2025-07-21)

### ♻️ Improvements

* Consolidate the configurations ([#13](https://github.com/inference-gateway/google-calendar-agent/issues/13)) ([f2aba96](https://github.com/inference-gateway/google-calendar-agent/commit/f2aba960c2dd34b9cc7a7c0127d4855ba8189673))
* Handle server creation error and update a2a dependency ([#14](https://github.com/inference-gateway/google-calendar-agent/issues/14)) ([c1b7c88](https://github.com/inference-gateway/google-calendar-agent/commit/c1b7c882efc7284dcf83a571e3d83e581d9664c8))

## [0.4.7](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.6...v0.4.7) (2025-07-20)

### ♻️ Improvements

* Add SERVER_DISABLE_HEALTH_LOGS configuration option to control health check logging ([b0e39db](https://github.com/inference-gateway/google-calendar-agent/commit/b0e39db450f86c2f764689201024012f3ca352c8))

## [0.4.6](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.5...v0.4.6) (2025-07-20)

### 🐛 Bug Fixes

* Consolidate ldflags for cleaner configuration in .goreleaser.yaml - keep it in one line ([8ff7b9c](https://github.com/inference-gateway/google-calendar-agent/commit/8ff7b9cd30e5e28817cb0293fef19417a058aebb))

## [0.4.5](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.4...v0.4.5) (2025-07-20)

### 🐛 Bug Fixes

* Update ldflags for improved build metadata and descriptions ([d84d652](https://github.com/inference-gateway/google-calendar-agent/commit/d84d652629fe14e2ca70c573ee524b60bc6a651f))

## [0.4.4](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.3...v0.4.4) (2025-07-20)

### 🐛 Bug Fixes

* Update ldflags to include versioning and agent details for Google Calendar Agent ([35b9688](https://github.com/inference-gateway/google-calendar-agent/commit/35b968856825c0435ae2232a8973e4b68e6a4310))

## [0.4.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.2...v0.4.3) (2025-07-20)

### ♻️ Improvements

* Replace local ldflags with library ldflags  ([#12](https://github.com/inference-gateway/google-calendar-agent/issues/12)) ([3256050](https://github.com/inference-gateway/google-calendar-agent/commit/325605094a1b86b56b23d651eee6fac991b4d3e3))

## [0.4.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.1...v0.4.2) (2025-07-19)

### 🐛 Bug Fixes

* Enable UPX compression for the built agent binary ([151b3a1](https://github.com/inference-gateway/google-calendar-agent/commit/151b3a1e178675e1d0b74e6bda1ebdae5eb7634e))
* Update dependencies for a2a and sdk to latest versions ([68c0b7c](https://github.com/inference-gateway/google-calendar-agent/commit/68c0b7c6b7012038f7c479bd9fcec3a212c00934))

## [0.4.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.4.0...v0.4.1) (2025-06-20)

### 🐛 Bug Fixes

* Improve the sed command to replace the hardcoded version when releasing ([b72d243](https://github.com/inference-gateway/google-calendar-agent/commit/b72d243a373cba9410f05c8815a1fc09f5107cad))

## [0.4.0](https://github.com/inference-gateway/google-calendar-agent/compare/v0.3.0...v0.4.0) (2025-06-20)

### ✨ Features

* Add AgentURL to configuration and update agent card URL ([0cdd79c](https://github.com/inference-gateway/google-calendar-agent/commit/0cdd79c7e5b6195f3a26e4dfdc5742cb3610d805))
* Add exec plugin for version update and include agent.go in release assets ([94fbca0](https://github.com/inference-gateway/google-calendar-agent/commit/94fbca0657e4b2c1acae1099b2f41aa9c781fa01))

## [0.3.0](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.6...v0.3.0) (2025-06-20)

### ✨ Features

* Add Streaming ([#11](https://github.com/inference-gateway/google-calendar-agent/issues/11)) ([1887fab](https://github.com/inference-gateway/google-calendar-agent/commit/1887fab0ce0009bc16b724e3f289da75ef340ba9))

### 📚 Documentation

* Example cleanup ([#10](https://github.com/inference-gateway/google-calendar-agent/issues/10)) ([398ef0d](https://github.com/inference-gateway/google-calendar-agent/commit/398ef0de716468e5fc009df7a778367408fd7031))

## [0.3.0-rc.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.3.0-rc.2...v0.3.0-rc.3) (2025-06-20)

### 🐛 Bug Fixes

* Update a2a dependency to v0.4.0-rc.10 ([c61de63](https://github.com/inference-gateway/google-calendar-agent/commit/c61de637259fb18e6cb4dbadb4c444986f729050))
* Update a2a dependency to v0.4.0-rc.8 and add streaming task submission example ([941810c](https://github.com/inference-gateway/google-calendar-agent/commit/941810c528d9ad0b8b681c83325637a5784bdf55))
* Update a2a dependency to v0.4.0-rc.9 ([1ba62be](https://github.com/inference-gateway/google-calendar-agent/commit/1ba62be567f6cabea644a83d9d46bd9ed5667200))

### 🔧 Miscellaneous

* Update inference-gateway and google-calendar-agent images to latest version ([3e9d1b5](https://github.com/inference-gateway/google-calendar-agent/commit/3e9d1b56607ef6530129391d3984858cd2f7873f))

## [0.3.0-rc.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.3.0-rc.1...v0.3.0-rc.2) (2025-06-19)

### 🐛 Bug Fixes

* Update a2a dependency to v0.4.0-rc.3 ([7f52aa3](https://github.com/inference-gateway/google-calendar-agent/commit/7f52aa39219510a162057bff77aa2f9a5f211b51))

## [0.3.0-rc.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.6...v0.3.0-rc.1) (2025-06-19)

### ✨ Features

* Update a2a dependency to v0.4.0-rc.1 ([091934e](https://github.com/inference-gateway/google-calendar-agent/commit/091934e2234d855a5161504f81dd1f8de652e047))

### 📚 Documentation

* Example cleanup ([#10](https://github.com/inference-gateway/google-calendar-agent/issues/10)) ([398ef0d](https://github.com/inference-gateway/google-calendar-agent/commit/398ef0de716468e5fc009df7a778367408fd7031))

## [0.2.6](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.5...v0.2.6) (2025-06-18)

### 🐛 Bug Fixes

* Refactor agent card retrieval and update dependencies ([fcae906](https://github.com/inference-gateway/google-calendar-agent/commit/fcae9065e7b873d39f073e3cd84ce78f2f8de617))

## [0.2.5](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.4...v0.2.5) (2025-06-17)

### 🐛 Bug Fixes

* Implement demo mode task handler with mock responses ([e1c9924](https://github.com/inference-gateway/google-calendar-agent/commit/e1c9924b6d16ebcc724e76970fc1ed7c5564bf1a))

## [0.2.4](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.3...v0.2.4) (2025-06-17)

### 🐛 Bug Fixes

* Correct main path for google-calendar-agent build ([8221f8b](https://github.com/inference-gateway/google-calendar-agent/commit/8221f8bb09760fdb286bb3951daa4d9313dd3177))
* Use agent only if we are not in demo mode ([cd1ec98](https://github.com/inference-gateway/google-calendar-agent/commit/cd1ec98fc90c1cb437ae63274094aef8ff5e4ce4))

## [0.2.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.2...v0.2.3) (2025-06-17)

### ♻️ Improvements

* Split giant file into multiple for clarity ([#9](https://github.com/inference-gateway/google-calendar-agent/issues/9)) ([96ea82d](https://github.com/inference-gateway/google-calendar-agent/commit/96ea82d407dc81466090d50190e7871b12806dc4))

## [0.2.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.1...v0.2.2) (2025-06-17)

### ♻️ Improvements

* Use Inference Gateway A2A ADK ([#6](https://github.com/inference-gateway/google-calendar-agent/issues/6)) ([fee77a8](https://github.com/inference-gateway/google-calendar-agent/commit/fee77a8084466daad960e245045f305399c4878e))

### 📚 Documentation

* Update README with enhanced configuration options and usage examples ([#8](https://github.com/inference-gateway/google-calendar-agent/issues/8)) ([7787deb](https://github.com/inference-gateway/google-calendar-agent/commit/7787deb9a9b2b2dc582b2ceeb62390282f353389))

### 🔧 Miscellaneous

* Remove TODOs related to A2A types and agent template ([6a335b5](https://github.com/inference-gateway/google-calendar-agent/commit/6a335b54ed2d568fee314eee368a40e65750f6e4))
* Update a2a dependency to stable version v0.2.0 ([#7](https://github.com/inference-gateway/google-calendar-agent/issues/7)) ([66cddd4](https://github.com/inference-gateway/google-calendar-agent/commit/66cddd461bec92e1fe75ee317677b620daa8d115))

## [0.2.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.0...v0.2.1) (2025-06-08)

### 📚 Documentation

* Cleanup a bit - remove internal service from direct interaction ([9aba7af](https://github.com/inference-gateway/google-calendar-agent/commit/9aba7af75f662675fa1470461a385d3c4c3ffe63))

### 🔧 Miscellaneous

* Update docker-compose.yaml ([7ffb366](https://github.com/inference-gateway/google-calendar-agent/commit/7ffb36677a5dcc2068dda24df6c3f523779ef578))

### 🔨 Miscellaneous

* Optimize Dockerfile.goreleaser by adding UPX compression for the agent ([#5](https://github.com/inference-gateway/google-calendar-agent/issues/5)) ([e7774fd](https://github.com/inference-gateway/google-calendar-agent/commit/e7774fd646450260913bb971d688fb2e597cfcd5))

## [0.2.0](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.3...v0.2.0) (2025-06-08)

### ✨ Features

* Enhance CalendarAgent with configuration support and timezone handling ([e7dd96c](https://github.com/inference-gateway/google-calendar-agent/commit/e7dd96c9bfa1b1e994c6e28e2881a6bdb63a4d4a))
* Implement Inference Gateway LLM Service ([#2](https://github.com/inference-gateway/google-calendar-agent/issues/2)) ([04eb229](https://github.com/inference-gateway/google-calendar-agent/commit/04eb22929e68ea2ee3221af771e80497a9c6bd23))

### 🔧 Miscellaneous

* Add timezone configuration for Google Calendar ([64bee3e](https://github.com/inference-gateway/google-calendar-agent/commit/64bee3ed3b1aadf235e36fa293bb5775efd155b9))

## [0.1.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.2...v0.1.3) (2025-06-08)

### ♻️ Improvements

* Update Google credentials handling and configuration options ([c953002](https://github.com/inference-gateway/google-calendar-agent/commit/c953002d318a940e9e421eb55f22292276adc99f))

## [0.1.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.1...v0.1.2) (2025-06-08)

### ♻️ Improvements

* Add configuration tests and utility functions for Google Calendar integration ([504b71c](https://github.com/inference-gateway/google-calendar-agent/commit/504b71c978bcb47d2592e492354e8c61d4f87f06))
* Introduce proper config ([#1](https://github.com/inference-gateway/google-calendar-agent/issues/1)) ([a34a9bd](https://github.com/inference-gateway/google-calendar-agent/commit/a34a9bdedfff096c93530d031e1f8d300b47fbb7))

### 🐛 Bug Fixes

* Correct environment variable name in logging and error messages for credentials file creation ([750c3be](https://github.com/inference-gateway/google-calendar-agent/commit/750c3becbffe10e0995b7f9952573b6096f6e08b))

### 📚 Documentation

* Add early development warning to README ([d597c28](https://github.com/inference-gateway/google-calendar-agent/commit/d597c28671154f2d7a005528b67d8744b62d836f))

### 🔧 Miscellaneous

* Improve cleanup task ([bc68e15](https://github.com/inference-gateway/google-calendar-agent/commit/bc68e15e5abb8b78eb3be44944afcba270c64ca5))

### ✅ Miscellaneous

* Add integration and mock tests for CalendarService functionality ([99b5ad2](https://github.com/inference-gateway/google-calendar-agent/commit/99b5ad23e49b6dfc3cf2fadf741407bad565e910))
* Enhance event request handling and add comprehensive tests for CalendarAgent ([7fcc2b0](https://github.com/inference-gateway/google-calendar-agent/commit/7fcc2b0527e9d0f6c37b791c2549a728f6d033ce))

## [0.1.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.0...v0.1.1) (2025-06-07)

### 🐛 Bug Fixes

* Correct counterfeiter version reference to v6.11.2 in Dockerfile, CI workflow, and goreleaser configuration ([a04daef](https://github.com/inference-gateway/google-calendar-agent/commit/a04daef076bc48dca26a2ef504b28be15890f05c))

### 🔧 Miscellaneous

* Update .gitignore to include /dist and add counterfeiter installation in .goreleaser.yaml ([610bbad](https://github.com/inference-gateway/google-calendar-agent/commit/610bbad23eb8792e039eecab2378af4585310d32))
* Update counterfeiter version to v6.11.2 in Dockerfile, CI workflow, and goreleaser configuration ([c58e832](https://github.com/inference-gateway/google-calendar-agent/commit/c58e832255df3557f1971bc42e0f669bd28653da))
