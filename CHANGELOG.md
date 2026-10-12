# Changelog

## [1.2.2](https://github.com/Apollogeddon/ignition-tofu/compare/v1.2.1...v1.2.2) (2026-10-11)


### Bug Fixes

* **client:** Escape names in request paths so ?, #, / and % stay in their segment ([2d8603a](https://github.com/Apollogeddon/ignition-tofu/commit/2d8603a11534331d458cf71250c3b552a6bac69a))
* **deps:** Override smol-toml to 1.9 past markdownlint-cli2's exact pin ([#82](https://github.com/Apollogeddon/ignition-tofu/issues/82)) ([f0c5d85](https://github.com/Apollogeddon/ignition-tofu/commit/f0c5d85449ae7187c7f7d6adb9758e1968be5222))
* Escape names in request paths, and destroy resources already deleted on the gateway ([4296d20](https://github.com/Apollogeddon/ignition-tofu/commit/4296d2035a545765dfa08a58fd42be19e967dbb5))
* **resources:** Treat a resource already deleted on the gateway as destroyed ([a857502](https://github.com/Apollogeddon/ignition-tofu/commit/a8575021ccd3ce21d133d0ec590af844f057c9de))
* **webpage:** Meet WCAG 2.2 AA and check the site's links and accessibility in CI ([#81](https://github.com/Apollogeddon/ignition-tofu/issues/81)) ([4e5e927](https://github.com/Apollogeddon/ignition-tofu/commit/4e5e9270f9d461bb72faa9a1413363d50a10dbef))

## [1.2.1](https://github.com/Apollogeddon/ignition-tofu/compare/v1.2.0...v1.2.1) (2026-10-10)


### Bug Fixes

* **deps:** Bump forgejs to 4.0.0 and let Dependabot update it ([5c841e7](https://github.com/Apollogeddon/ignition-tofu/commit/5c841e78275b0d93ce0450bf6ead845649c5067d))
* **deps:** Bump forgejs to 4.0.0, which bundles with tsdown instead of tsup ([7e8696d](https://github.com/Apollogeddon/ignition-tofu/commit/7e8696de29abadeaa6bf8abe1a185f0359b5db7b))
* **deps:** Override katex to a patched version ([#66](https://github.com/Apollogeddon/ignition-tofu/issues/66)) ([2ea036f](https://github.com/Apollogeddon/ignition-tofu/commit/2ea036f0569104d33310805d2365ef5553d00089))

## [1.2.0](https://github.com/Apollogeddon/ignition-tfpl/compare/v1.1.1...v1.2.0) (2026-10-09)


### Features

* Serve the provider from a network mirror on the docs site ([793987e](https://github.com/Apollogeddon/ignition-tfpl/commit/793987e50f6939abbd5d1bd55d6ce0ec611e6c1c))
* Serve the provider from a network mirror on the docs site ([e756888](https://github.com/Apollogeddon/ignition-tfpl/commit/e75688861330da73bb7a392ba8fe369fdf538b15))

## [1.1.1](https://github.com/Apollogeddon/ignition-tfpl/compare/v1.1.0...v1.1.1) (2026-10-09)


### Bug Fixes

* **deps:** Upgrade modules with known vulnerabilities via govulncheck ([70cfed7](https://github.com/Apollogeddon/ignition-tfpl/commit/70cfed723bc9914416e7cab518ba1dbad7f4526d))
* **resources:** Remove resources deleted outside OpenTofu from state on read ([2ec5222](https://github.com/Apollogeddon/ignition-tfpl/commit/2ec52226e86433cede82fa77fd87f44fd8a7bb47))

## [1.1.0](https://github.com/Apollogeddon/ignition-tfpl/compare/v1.0.0...v1.1.0) (2026-10-08)


### Features

* Release the provider in the Terraform Registry's format ([bfee84b](https://github.com/Apollogeddon/ignition-tfpl/commit/bfee84bef69dbefc793f66ec5fb0e9a841b2e9ff))


### Bug Fixes

* **client:** Fix error handling, secret retries, and identity-provider REST payload ([8a670cf](https://github.com/Apollogeddon/ignition-tfpl/commit/8a670cfde69ca299bba69c857549b7a57d0a23bc))
* **client:** Fix redundancy, gan-settings, and device REST API mismatches ([c1ea571](https://github.com/Apollogeddon/ignition-tfpl/commit/c1ea5713e52c0e25befc8fba8b3155301d67898b))
* **deps:** Upgrade hc-install and modules with known vulnerabilities ([9376ecb](https://github.com/Apollogeddon/ignition-tfpl/commit/9376ecb20550d3d17cd7313d6e2b3e3933cd2c18))
* **docker:** Build gwbk into gateway image so restore works on remote docker hosts ([1fa30b9](https://github.com/Apollogeddon/ignition-tfpl/commit/1fa30b928acc6ed8aa3659167c5f4d6f56b40e2c))
* **resources:** Correct module and type sent for alarm-journal resource ([2465b98](https://github.com/Apollogeddon/ignition-tfpl/commit/2465b987b0add7947a0c09d979e4ed51f94a2785))
* **resources:** Default opc-ua connection type to avoid sending an empty profile type ([8aa2c97](https://github.com/Apollogeddon/ignition-tfpl/commit/8aa2c9766d392edc9e9277d5c43c13ed642ee8a4))

## 1.0.0 (2026-01-26)


### Features

* **alarm-journal:** add alarm journal resource ([d856142](https://github.com/Apollogeddon/ignition-tfpl/commit/d856142e8d6a35d7cfae577875f9e47500bcdd00))
* **alarm-notification-profile:** add alarm notification profile resource ([fa26f3c](https://github.com/Apollogeddon/ignition-tfpl/commit/fa26f3c8a361b2b4530b4035a2e44c5e2fc73f5b))
* **audit-profile:** add audit profile resource ([1c91c5d](https://github.com/Apollogeddon/ignition-tfpl/commit/1c91c5d1144b334dd4547a555afadb3aab4c3135))
* **client, provider:** enhance resource management and client testing ([3539e02](https://github.com/Apollogeddon/ignition-tfpl/commit/3539e02671b9f6f2533aa6a1f6cfe7e6c32c29f3))
* **client:** increase retry attempts and add comprehensive tests ([6fe27ee](https://github.com/Apollogeddon/ignition-tfpl/commit/6fe27eec259b53ec065fa3ccd5d30e4f56cd141d))
* **client:** introduce Ignition API client package ([96bc6f8](https://github.com/Apollogeddon/ignition-tfpl/commit/96bc6f83e3b2ca6aa1f338b606de75df0aaf78ee))
* **db-connection:** introduce database connection resource and data source ([e36e173](https://github.com/Apollogeddon/ignition-tfpl/commit/e36e1730e28298fbfba5ee99dbfb925c0239ef51))
* **device:** implement Ignition device resource ([18be8ee](https://github.com/Apollogeddon/ignition-tfpl/commit/18be8eeb23c10eabd2349f7bb8875e18852f357b))
* **docs, client:** add usage examples and enhance error handling ([e87cb5e](https://github.com/Apollogeddon/ignition-tfpl/commit/e87cb5e69a48660ce7d396f60ac30cf9930abbba))
* **docs:** add documentation website and CI deployment ([802e302](https://github.com/Apollogeddon/ignition-tfpl/commit/802e30280e9fa25c7b167ca3873e435cbee2682f))
* **docs:** add script to migrate and transform documentation ([21afdc8](https://github.com/Apollogeddon/ignition-tfpl/commit/21afdc86e2bcc179e10544ee7ed76bc64260a708))
* **examples:** add Ignition provider resource & data source examples ([40e2bfe](https://github.com/Apollogeddon/ignition-tfpl/commit/40e2bfeeccc2f604997ac9059f147ac7b561018f))
* **idp:** add identity provider resource ([e78470d](https://github.com/Apollogeddon/ignition-tfpl/commit/e78470dc8e581005685869801346d43855ba54e3))
* **idp:** add oidc and saml identity provider types ([6bc8f8d](https://github.com/Apollogeddon/ignition-tfpl/commit/6bc8f8d6cc785985aa9200de74b14fe7ec0bfac1))
* **opc-connection:** add OPC UA connection resource ([3094606](https://github.com/Apollogeddon/ignition-tfpl/commit/3094606ff285ebee599d4cd9d0d96234980f667e))
* **project:** add project resource and data source ([3be0903](https://github.com/Apollogeddon/ignition-tfpl/commit/3be090369ba835cc5f58b6c54352f5498d8e1fcb))
* **provider-gan:** add GAN outgoing connection resource ([8c50daf](https://github.com/Apollogeddon/ignition-tfpl/commit/8c50daf8290679a3f6f4f9d01fe4534b86b2e11d))
* **provider-gan:** add general gateway network settings resource ([4fb7727](https://github.com/Apollogeddon/ignition-tfpl/commit/4fb77273f88730877e705368030cc72688e6a309))
* **provider:** implement core provider structure and configuration ([85b0fb8](https://github.com/Apollogeddon/ignition-tfpl/commit/85b0fb850ca7e86f0430a85bd976b927c87b22eb))
* **provider:** implement secret encryption and enhance resource handling ([7d33ce6](https://github.com/Apollogeddon/ignition-tfpl/commit/7d33ce6f357385f25a3078ee0c62e396e91d53ca))
* **provider:** initialise ignition Terraform provider ([e109e0d](https://github.com/Apollogeddon/ignition-tfpl/commit/e109e0d3b48c5f70bced6fa0ac902bbfce334fb9))
* **provider:** introduce generic resource handler and new resources ([2b7abae](https://github.com/Apollogeddon/ignition-tfpl/commit/2b7abae6afae8390164b4bbd41b93df109e20b5e))
* **resources:** add client and provider for several types ([140618e](https://github.com/Apollogeddon/ignition-tfpl/commit/140618e7c9fd20bf6cde541d1ca75b654f7db378))
* **smtp-profile:** add SMTP profile resource and data source ([3444eaf](https://github.com/Apollogeddon/ignition-tfpl/commit/3444eafa8bea9625eebc2fb6ed8bff16c25dc252))
* **store-forward:** add store and forward engine resource and data source ([b2264ff](https://github.com/Apollogeddon/ignition-tfpl/commit/b2264ff60d158ca50b53da8127f9e15c6aefe3d9))
* **tag_provider:** add settings field and standardise type casing ([0136aaf](https://github.com/Apollogeddon/ignition-tfpl/commit/0136aaf1885c740daf138f09dca7229106cce199))
* **tag-provider:** add tag provider resource and data source ([8c6b909](https://github.com/Apollogeddon/ignition-tfpl/commit/8c6b9091aa7cf20beed88e71e398a469d8afaf32))
* **user-source:** add user source resource and data source ([22b31ca](https://github.com/Apollogeddon/ignition-tfpl/commit/22b31ca8b98adbcea59dfff5556f52c93c6ba502))
* **webpage:** enhance landing page and update base path ([e3c7248](https://github.com/Apollogeddon/ignition-tfpl/commit/e3c724885c737d03220aa2679a8b38f78d70447f))
