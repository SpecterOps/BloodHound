package typescript

type DependencyMap map[string]string

type Foo struct {
	Dependencies        Foo_sub317        `json:"dependencies"`
	PrivateDependencies DependencyMap     `json:"_dependencies"`
	DevDependencies     DependencyMap     `json:"devDependencies"`
	Extraneous          bool              `json:"extraneous"`
	Name                string            `json:"name"`
	PackageManager      string            `json:"packageManager"`
	Path                string            `json:"path"`
	PeerDependencies    DependencyMap     `json:"peerDependencies"`
	Private             bool              `json:"private"`
	Problems            []string          `json:"problems"`
	Scripts             map[string]string `json:"scripts"`
	Workspaces          []string          `json:"workspaces"`
}

type Foo_sub9 struct {
	Access string `json:"access"`
}

type Foo_sub73 struct {
	Add_contributor string `json:"add-contributor"`
	Build           string `json:"build"`
	Commit          string `json:"commit"`
	Dev             string `json:"dev"`
	Lint            string `json:"lint"`
	Start           string `json:"start"`
	Test            string `json:"test"`
	Validate        string `json:"validate"`
}

type Foo_sub199 struct {
	All       bool     `json:"all"`
	Exclude   []string `json:"exclude"`
	Extension []string `json:"extension"`
	Include   []string `json:"include"`
	Reporter  []string `json:"reporter"`
}

type Foo_sub13 struct {
	All_contributors string `json:"all-contributors"`
}

type Foo_sub194 struct {
	Angular_devkit_build_optimizer     string `json:"@angular-devkit/build-optimizer"`
	Angular_devkit_schematics          string `json:"@angular-devkit/schematics"`
	Types_chai                         string `json:"@types/chai"`
	Types_lodash                       string `json:"@types/lodash"`
	Types_mocha                        string `json:"@types/mocha"`
	Types_node                         string `json:"@types/node"`
	Types_sinon                        string `json:"@types/sinon"`
	Types_sinon_chai                   string `json:"@types/sinon-chai"`
	Types_source_map                   string `json:"@types/source-map"`
	Babel_polyfill                     string `json:"babel-polyfill"`
	Benchmark                          string `json:"benchmark"`
	Benchpress                         string `json:"benchpress"`
	Chai                               string `json:"chai"`
	Check_side_effects                 string `json:"check-side-effects"`
	Color                              string `json:"color"`
	Colors                             string `json:"colors"`
	Commitizen                         string `json:"commitizen"`
	Coveralls                          string `json:"coveralls"`
	Cross_env                          string `json:"cross-env"`
	Cz_conventional_changelog          string `json:"cz-conventional-changelog"`
	Danger                             string `json:"danger"`
	Dependency_cruiser                 string `json:"dependency-cruiser"`
	Doctoc                             string `json:"doctoc"`
	Dtslint                            string `json:"dtslint"`
	Escape_string_regexp               string `json:"escape-string-regexp"`
	Esdoc                              string `json:"esdoc"`
	Eslint                             string `json:"eslint"`
	Eslint_plugin_jasmine              string `json:"eslint-plugin-jasmine"`
	Fs_extra                           string `json:"fs-extra"`
	Get_folder_size                    string `json:"get-folder-size"`
	Glob                               string `json:"glob"`
	Gm                                 string `json:"gm"`
	Google_closure_compiler_js         string `json:"google-closure-compiler-js"`
	Gzip_size                          string `json:"gzip-size"`
	HTTP_server                        string `json:"http-server"`
	Husky                              string `json:"husky"`
	Klaw_sync                          string `json:"klaw-sync"`
	Lint_staged                        string `json:"lint-staged"`
	Lodash                             string `json:"lodash"`
	Markdown_doctest                   string `json:"markdown-doctest"`
	Minimist                           string `json:"minimist"`
	Mkdirp                             string `json:"mkdirp"`
	Mocha                              string `json:"mocha"`
	Mocha_in_sauce                     string `json:"mocha-in-sauce"`
	Npm_run_all                        string `json:"npm-run-all"`
	Nyc                                string `json:"nyc"`
	Opn_cli                            string `json:"opn-cli"`
	Platform                           string `json:"platform"`
	Promise                            string `json:"promise"`
	Protractor                         string `json:"protractor"`
	Rollup                             string `json:"rollup"`
	Rollup_plugin_alias                string `json:"rollup-plugin-alias"`
	Rollup_plugin_inject               string `json:"rollup-plugin-inject"`
	Rollup_plugin_node_resolve         string `json:"rollup-plugin-node-resolve"`
	Rx                                 string `json:"rx"`
	Rxjs                               string `json:"rxjs"`
	Shx                                string `json:"shx"`
	Sinon                              string `json:"sinon"`
	Sinon_chai                         string `json:"sinon-chai"`
	Source_map_support                 string `json:"source-map-support"`
	Symbol_observable                  string `json:"symbol-observable"`
	Systemjs                           string `json:"systemjs"`
	Ts_api_guardian                    string `json:"ts-api-guardian"`
	Ts_node                            string `json:"ts-node"`
	Tsconfig_paths                     string `json:"tsconfig-paths"`
	Tslint                             string `json:"tslint"`
	Tslint_etc                         string `json:"tslint-etc"`
	Tslint_no_toplevel_property_access string `json:"tslint-no-toplevel-property-access"`
	Tslint_no_unused_expression_chai   string `json:"tslint-no-unused-expression-chai"`
	Typescript                         string `json:"typescript"`
	Validate_commit_msg                string `json:"validate-commit-msg"`
	Webpack                            string `json:"webpack"`
	Xmlhttprequest                     string `json:"xmlhttprequest"`
}

type Foo_sub204 struct {
	Ansi_escapes    Foo_sub184 `json:"ansi-escapes"`
	Chalk           Foo_sub27  `json:"chalk"`
	Cli_cursor      Foo_sub185 `json:"cli-cursor"`
	Cli_width       Foo_sub186 `json:"cli-width"`
	External_editor Foo_sub187 `json:"external-editor"`
	Figures         Foo_sub188 `json:"figures"`
	Lodash          Foo_sub47  `json:"lodash"`
	Mute_stream     Foo_sub190 `json:"mute-stream"`
	Run_async       Foo_sub192 `json:"run-async"`
	Rxjs            Foo_sub200 `json:"rxjs"`
	String_width    Foo_sub124 `json:"string-width"`
	Strip_ansi      Foo_sub126 `json:"strip-ansi"`
	Through         Foo_sub203 `json:"through"`
}

type Foo_sub30 struct {
	Ansi_escapes    string `json:"ansi-escapes"`
	Chalk           string `json:"chalk"`
	Cli_cursor      string `json:"cli-cursor"`
	Cli_width       string `json:"cli-width"`
	External_editor string `json:"external-editor"`
	Figures         string `json:"figures"`
	Lodash          string `json:"lodash"`
	Mute_stream     string `json:"mute-stream"`
	Run_async       string `json:"run-async"`
	Rxjs            string `json:"rxjs"`
	String_width    string `json:"string-width"`
	Strip_ansi      string `json:"strip-ansi"`
	Through         string `json:"through"`
}

type Foo_sub269 struct {
	Ansi_regex Foo_sub268 `json:"ansi-regex"`
}

type Foo_sub125 struct {
	Ansi_regex string `json:"ansi-regex"`
}

type Foo_sub297 struct {
	Ansi_styles  Foo_sub101 `json:"ansi-styles"`
	String_width Foo_sub124 `json:"string-width"`
	Strip_ansi   Foo_sub126 `json:"strip-ansi"`
}

type Foo_sub105 struct {
	Ansi_styles    Foo_sub101 `json:"ansi-styles"`
	Supports_color Foo_sub104 `json:"supports-color"`
}

type Foo_sub127 struct {
	Ansi_styles  string `json:"ansi-styles"`
	String_width string `json:"string-width"`
	Strip_ansi   string `json:"strip-ansi"`
}

type Foo_sub22 struct {
	Ansi_styles    string `json:"ansi-styles"`
	Supports_color string `json:"supports-color"`
}

type Foo_sub152 struct {
	Async    string `json:"async"`
	Errto    string `json:"errto"`
	Iconv    string `json:"iconv"`
	Istanbul string `json:"istanbul"`
	Mocha    string `json:"mocha"`
	Request  string `json:"request"`
	Semver   string `json:"semver"`
	Unorm    string `json:"unorm"`
}

type Foo_sub128 struct {
	Ava       string `json:"ava"`
	Chalk     string `json:"chalk"`
	Coveralls string `json:"coveralls"`
	Has_ansi  string `json:"has-ansi"`
	Nyc       string `json:"nyc"`
	Xo        string `json:"xo"`
}

type Foo_sub23 struct {
	Ava          string `json:"ava"`
	Coveralls    string `json:"coveralls"`
	Execa        string `json:"execa"`
	Import_fresh string `json:"import-fresh"`
	Matcha       string `json:"matcha"`
	Nyc          string `json:"nyc"`
	Resolve_from string `json:"resolve-from"`
	Tsd          string `json:"tsd"`
	Xo           string `json:"xo"`
}

type Foo_sub294 struct {
	Ava              string `json:"ava"`
	Coveralls        string `json:"coveralls"`
	Nyc              string `json:"nyc"`
	Standard         string `json:"standard"`
	Standard_version string `json:"standard-version"`
}

type Foo_sub231 struct {
	Ava        string `json:"ava"`
	Delay      string `json:"delay"`
	In_range   string `json:"in-range"`
	Random_int string `json:"random-int"`
	Time_span  string `json:"time-span"`
	Tsd_check  string `json:"tsd-check"`
	Xo         string `json:"xo"`
}

type Foo_sub214 struct {
	Ava       string `json:"ava"`
	Delay     string `json:"delay"`
	In_range  string `json:"in-range"`
	Time_span string `json:"time-span"`
	Tsd       string `json:"tsd"`
	Xo        string `json:"xo"`
}

type Foo_sub103 struct {
	Ava          string `json:"ava"`
	Import_fresh string `json:"import-fresh"`
	Xo           string `json:"xo"`
}

type Foo_sub174 struct {
	Ava            string `json:"ava"`
	Is_path_inside string `json:"is-path-inside"`
	Tempy          string `json:"tempy"`
	Tsd            string `json:"tsd"`
	Xo             string `json:"xo"`
}

type Foo_sub166 struct {
	Ava            string `json:"ava"`
	Markdown_table string `json:"markdown-table"`
	Tsd            string `json:"tsd"`
	Xo             string `json:"xo"`
}

type Foo_sub56 struct {
	Ava            string `json:"ava"`
	Pinkie_promise string `json:"pinkie-promise"`
	V8_natives     string `json:"v8-natives"`
	Xo             string `json:"xo"`
}

type Foo_sub85 struct {
	Ava string `json:"ava"`
	Tsd string `json:"tsd"`
	Xo  string `json:"xo"`
}

type Foo_sub206 struct {
	Ava       string `json:"ava"`
	Tsd_check string `json:"tsd-check"`
	Xo        string `json:"xo"`
}

type Foo_sub139 struct {
	Ava string `json:"ava"`
	Xo  string `json:"xo"`
}

type Foo_sub142 struct {
	Babel_cli                                    string `json:"@babel/cli"`
	Babel_core                                   string `json:"@babel/core"`
	Babel_plugin_proposal_unicode_property_regex string `json:"@babel/plugin-proposal-unicode-property-regex"`
	Babel_preset_env                             string `json:"@babel/preset-env"`
	Mocha                                        string `json:"mocha"`
	Regexgen                                     string `json:"regexgen"`
	Unicode_12_0_0                               string `json:"unicode-12.0.0"`
}

type Foo_sub18 struct {
	Babel_eslint_parser                            string `json:"@babel/eslint-parser"`
	Babel_core                                     string `json:"babel-core"`
	Babel_minify                                   string `json:"babel-minify"`
	Babel_plugin_add_module_exports                string `json:"babel-plugin-add-module-exports"`
	Babel_plugin_istanbul                          string `json:"babel-plugin-istanbul"`
	Babel_plugin_syntax_async_generators           string `json:"babel-plugin-syntax-async-generators"`
	Babel_plugin_transform_es2015_modules_commonjs string `json:"babel-plugin-transform-es2015-modules-commonjs"`
	Babel_preset_es2015                            string `json:"babel-preset-es2015"`
	Babel_preset_es2017                            string `json:"babel-preset-es2017"`
	Babel_register                                 string `json:"babel-register"`
	Babelify                                       string `json:"babelify"`
	Benchmark                                      string `json:"benchmark"`
	Bluebird                                       string `json:"bluebird"`
	Browserify                                     string `json:"browserify"`
	Chai                                           string `json:"chai"`
	Cheerio                                        string `json:"cheerio"`
	Es6_promise                                    string `json:"es6-promise"`
	Eslint                                         string `json:"eslint"`
	Eslint_plugin_prefer_arrow                     string `json:"eslint-plugin-prefer-arrow"`
	Fs_extra                                       string `json:"fs-extra"`
	Jsdoc                                          string `json:"jsdoc"`
	Karma                                          string `json:"karma"`
	Karma_browserify                               string `json:"karma-browserify"`
	Karma_firefox_launcher                         string `json:"karma-firefox-launcher"`
	Karma_mocha                                    string `json:"karma-mocha"`
	Karma_mocha_reporter                           string `json:"karma-mocha-reporter"`
	Karma_safari_launcher                          string `json:"karma-safari-launcher"`
	Mocha                                          string `json:"mocha"`
	Native_promise_only                            string `json:"native-promise-only"`
	Nyc                                            string `json:"nyc"`
	Rollup                                         string `json:"rollup"`
	Rollup_plugin_node_resolve                     string `json:"rollup-plugin-node-resolve"`
	Rollup_plugin_npm                              string `json:"rollup-plugin-npm"`
	Rsvp                                           string `json:"rsvp"`
	Semver                                         string `json:"semver"`
	Yargs                                          string `json:"yargs"`
}

type Foo_sub317 struct {
	Babel_runtime           Foo_sub11  `json:"@babel/runtime"`
	All_contributors_cli    Foo_sub74  `json:"all-contributors-cli"`
	Ansi_escapes            Foo_sub84  `json:"ansi-escapes"`
	Ansi_regex              Foo_sub87  `json:"ansi-regex"`
	Ansi_styles             Foo_sub98  `json:"ansi-styles"`
	Async                   Foo_sub99  `json:"async"`
	Camelcase               Foo_sub100 `json:"camelcase"`
	Chalk                   Foo_sub106 `json:"chalk"`
	Chardet                 Foo_sub110 `json:"chardet"`
	Cli_cursor              Foo_sub116 `json:"cli-cursor"`
	Cli_width               Foo_sub119 `json:"cli-width"`
	Cliui                   Foo_sub134 `json:"cliui"`
	Color_convert           Foo_sub137 `json:"color-convert"`
	Color_name              Foo_sub138 `json:"color-name"`
	Decamelize              Foo_sub140 `json:"decamelize"`
	Didyoumean              Foo_sub141 `json:"didyoumean"`
	Emoji_regex             Foo_sub144 `json:"emoji-regex"`
	Escape_string_regexp    Foo_sub145 `json:"escape-string-regexp"`
	External_editor         Foo_sub162 `json:"external-editor"`
	Figures                 Foo_sub168 `json:"figures"`
	Find_up                 Foo_sub175 `json:"find-up"`
	Get_caller_file         Foo_sub178 `json:"get-caller-file"`
	Has_flag                Foo_sub100 `json:"has-flag"`
	Iconv_lite              Foo_sub183 `json:"iconv-lite"`
	Inquirer                Foo_sub205 `json:"inquirer"`
	Is_fullwidth_code_point Foo_sub207 `json:"is-fullwidth-code-point"`
	JSON_fixer              Foo_sub212 `json:"json-fixer"`
	Locate_path             Foo_sub217 `json:"locate-path"`
	Lodash                  Foo_sub218 `json:"lodash"`
	Mimic_fn                Foo_sub100 `json:"mimic-fn"`
	Mute_stream             Foo_sub219 `json:"mute-stream"`
	Node_fetch              Foo_sub225 `json:"node-fetch"`
	Onetime                 Foo_sub228 `json:"onetime"`
	Os_tmpdir               Foo_sub140 `json:"os-tmpdir"`
	P_limit                 Foo_sub232 `json:"p-limit"`
	P_locate                Foo_sub235 `json:"p-locate"`
	P_try                   Foo_sub100 `json:"p-try"`
	Path_exists             Foo_sub100 `json:"path-exists"`
	Pegjs                   Foo_sub236 `json:"pegjs"`
	Pify                    Foo_sub237 `json:"pify"`
	Prettier                Foo_sub238 `json:"prettier"`
	Regenerator_runtime     Foo_sub239 `json:"regenerator-runtime"`
	Require_directory       Foo_sub243 `json:"require-directory"`
	Require_main_filename   Foo_sub246 `json:"require-main-filename"`
	Restore_cursor          Foo_sub252 `json:"restore-cursor"`
	Run_async               Foo_sub253 `json:"run-async"`
	Rxjs                    Foo_sub258 `json:"rxjs"`
	Safer_buffer            Foo_sub259 `json:"safer-buffer"`
	Set_blocking            Foo_sub262 `json:"set-blocking"`
	Signal_exit             Foo_sub263 `json:"signal-exit"`
	String_width            Foo_sub267 `json:"string-width"`
	Strip_ansi              Foo_sub270 `json:"strip-ansi"`
	Supports_color          Foo_sub272 `json:"supports-color"`
	Through                 Foo_sub273 `json:"through"`
	Tmp                     Foo_sub276 `json:"tmp"`
	Tr46                    Foo_sub279 `json:"tr46"`
	Tslib                   Foo_sub280 `json:"tslib"`
	Type_fest               Foo_sub281 `json:"type-fest"`
	Typescript              Foo_sub288 `json:"typescript"`
	Webidl_conversions      Foo_sub289 `json:"webidl-conversions"`
	Whatwg_url              Foo_sub293 `json:"whatwg-url"`
	Which_module            Foo_sub296 `json:"which-module"`
	Wrap_ansi               Foo_sub298 `json:"wrap-ansi"`
	Y18n                    Foo_sub300 `json:"y18n"`
	Yargs                   Foo_sub314 `json:"yargs"`
	Yargs_parser            Foo_sub316 `json:"yargs-parser"`
}

type Foo_sub67 struct {
	Babel_runtime Foo_sub17 `json:"@babel/runtime"`
	Async         Foo_sub21 `json:"async"`
	Chalk         Foo_sub27 `json:"chalk"`
	Didyoumean    Foo_sub29 `json:"didyoumean"`
	Inquirer      Foo_sub33 `json:"inquirer"`
	JSON_fixer    Foo_sub45 `json:"json-fixer"`
	Lodash        Foo_sub47 `json:"lodash"`
	Node_fetch    Foo_sub55 `json:"node-fetch"`
	Pify          Foo_sub58 `json:"pify"`
	Prettier      Foo_sub60 `json:"prettier"`
	Yargs         Foo_sub66 `json:"yargs"`
}

type Foo_sub211 struct {
	Babel_runtime Foo_sub17  `json:"@babel/runtime"`
	Chalk         Foo_sub27  `json:"chalk"`
	Pegjs         Foo_sub210 `json:"pegjs"`
}

type Foo_sub12 struct {
	Babel_runtime string `json:"@babel/runtime"`
	Async         string `json:"async"`
	Chalk         string `json:"chalk"`
	Didyoumean    string `json:"didyoumean"`
	Inquirer      string `json:"inquirer"`
	JSON_fixer    string `json:"json-fixer"`
	Lodash        string `json:"lodash"`
	Node_fetch    string `json:"node-fetch"`
	Pify          string `json:"pify"`
	Prettier      string `json:"prettier"`
	Yargs         string `json:"yargs"`
}

type Foo_sub34 struct {
	Babel_runtime string `json:"@babel/runtime"`
	Chalk         string `json:"chalk"`
	Pegjs         string `json:"pegjs"`
}

type Foo_sub24 struct {
	Bench string `json:"bench"`
	Test  string `json:"test"`
}

type Foo_sub122 struct {
	Blanket Foo_sub121 `json:"blanket"`
}

type Foo_sub53 struct {
	Branches []interface{} `json:"branches"`
}

type Foo_sub40 struct {
	Branches   int64 `json:"branches"`
	Functions  int64 `json:"functions"`
	Lines      int64 `json:"lines"`
	Statements int64 `json:"statements"`
}

type Foo_sub209 struct {
	Browserify   string `json:"browserify"`
	Eslint       string `json:"eslint"`
	HTTP_server  string `json:"http-server"`
	Jasmine_node string `json:"jasmine-node"`
	Uglify_js    string `json:"uglify-js"`
}

type Foo_sub180 struct {
	Browserify_test string `json:"browserify-test"`
	Test            string `json:"test"`
}

type Foo_sub202 struct {
	Browsers []string `json:"browsers"`
	Files    string   `json:"files"`
}

type Foo_sub286 struct {
	Build                   string `json:"build"`
	Build_compiler          string `json:"build:compiler"`
	Build_tests             string `json:"build:tests"`
	Build_tests_notypecheck string `json:"build:tests:notypecheck"`
	Clean                   string `json:"clean"`
	Gulp                    string `json:"gulp"`
	Lint                    string `json:"lint"`
	Setup_hooks             string `json:"setup-hooks"`
	Start                   string `json:"start"`
	Test                    string `json:"test"`
	Test_eslint_rules       string `json:"test:eslint-rules"`
}

type Foo_sub318 struct {
	Build        string `json:"build"`
	Check_format string `json:"check-format"`
	Check_types  string `json:"check-types"`
	Debug        string `json:"debug"`
	Dev          string `json:"dev"`
	Format       string `json:"format"`
	Lint         string `json:"lint"`
	Preview      string `json:"preview"`
	Start        string `json:"start"`
	Test         string `json:"test"`
}

type Foo_sub44 struct {
	Build         string `json:"build"`
	Commit        string `json:"commit"`
	Format        string `json:"format"`
	Generate      string `json:"generate"`
	Lint          string `json:"lint"`
	Lint_fix      string `json:"lint:fix"`
	Lint_lockfile string `json:"lint:lockfile"`
	Lint_md       string `json:"lint:md"`
	Prepare       string `json:"prepare"`
	Sandbox       string `json:"sandbox"`
	Sec           string `json:"sec"`
	Sr            string `json:"sr"`
	Test          string `json:"test"`
	Test_watch    string `json:"test:watch"`
}

type Foo_sub222 struct {
	Build      string `json:"build"`
	Coverage   string `json:"coverage"`
	Lint       string `json:"lint"`
	Prepublish string `json:"prepublish"`
	Pretest    string `json:"pretest"`
	Test       string `json:"test"`
}

type Foo_sub54 struct {
	Build    string `json:"build"`
	Coverage string `json:"coverage"`
	Prepare  string `json:"prepare"`
	Report   string `json:"report"`
	Test     string `json:"test"`
}

type Foo_sub143 struct {
	Build      string `json:"build"`
	Test       string `json:"test"`
	Test_watch string `json:"test:watch"`
}

type Foo_sub310 struct {
	C8       string `json:"c8"`
	Chai     string `json:"chai"`
	Mocha    string `json:"mocha"`
	Standard string `json:"standard"`
}

type Foo_sub315 struct {
	Camelcase  Foo_sub172 `json:"camelcase"`
	Decamelize Foo_sub274 `json:"decamelize"`
}

type Foo_sub69 struct {
	Camelcase         string `json:"camelcase"`
	Consistent_return string `json:"consistent-return"`
	Func_names        string `json:"func-names"`
	Import_extensions string `json:"import/extensions"`
	No_process_exit   string `json:"no-process-exit"`
}

type Foo_sub309 struct {
	Camelcase  string `json:"camelcase"`
	Decamelize string `json:"decamelize"`
}

type Foo_sub131 struct {
	Chai      string `json:"chai"`
	Chalk     string `json:"chalk"`
	Coveralls string `json:"coveralls"`
	Mocha     string `json:"mocha"`
	Nyc       string `json:"nyc"`
	Standard  string `json:"standard"`
}

type Foo_sub31 struct {
	Chai       string `json:"chai"`
	Chalk_pipe string `json:"chalk-pipe"`
	Cmdify     string `json:"cmdify"`
	Mocha      string `json:"mocha"`
	Mockery    string `json:"mockery"`
	Nyc        string `json:"nyc"`
	Sinon      string `json:"sinon"`
}

type Foo_sub299 struct {
	Chai             string `json:"chai"`
	Coveralls        string `json:"coveralls"`
	Mocha            string `json:"mocha"`
	Nyc              string `json:"nyc"`
	Rimraf           string `json:"rimraf"`
	Standard         string `json:"standard"`
	Standard_version string `json:"standard-version"`
}

type Foo_sub260 struct {
	Chai             string `json:"chai"`
	Coveralls        string `json:"coveralls"`
	Mocha            string `json:"mocha"`
	Nyc              string `json:"nyc"`
	Standard         string `json:"standard"`
	Standard_version string `json:"standard-version"`
}

type Foo_sub248 struct {
	Chai             string `json:"chai"`
	Coveralls        string `json:"coveralls"`
	Nyc              string `json:"nyc"`
	Standard_version string `json:"standard-version"`
	Tap              string `json:"tap"`
}

type Foo_sub244 struct {
	Chai             string `json:"chai"`
	Standard         string `json:"standard"`
	Standard_version string `json:"standard-version"`
	Tap              string `json:"tap"`
}

type Foo_sub90 struct {
	Chalk string `json:"chalk"`
	Xo    string `json:"xo"`
}

type Foo_sub159 struct {
	Chardet    Foo_sub149 `json:"chardet"`
	Iconv_lite Foo_sub154 `json:"iconv-lite"`
	Tmp        Foo_sub158 `json:"tmp"`
}

type Foo_sub146 struct {
	Chardet    string `json:"chardet"`
	Iconv_lite string `json:"iconv-lite"`
	Tmp        string `json:"tmp"`
}

type Foo_sub64 struct {
	Check    string `json:"check"`
	Compile  string `json:"compile"`
	Coverage string `json:"coverage"`
	Fix      string `json:"fix"`
	Posttest string `json:"posttest"`
	Prepare  string `json:"prepare"`
	Pretest  string `json:"pretest"`
	Test     string `json:"test"`
}

type Foo_sub313 struct {
	Cliui                 Foo_sub301 `json:"cliui"`
	Decamelize            Foo_sub274 `json:"decamelize"`
	Find_up               Foo_sub302 `json:"find-up"`
	Get_caller_file       Foo_sub303 `json:"get-caller-file"`
	Require_directory     Foo_sub304 `json:"require-directory"`
	Require_main_filename Foo_sub305 `json:"require-main-filename"`
	Set_blocking          Foo_sub306 `json:"set-blocking"`
	String_width          Foo_sub124 `json:"string-width"`
	Which_module          Foo_sub307 `json:"which-module"`
	Y18n                  Foo_sub308 `json:"y18n"`
	Yargs_parser          Foo_sub312 `json:"yargs-parser"`
}

type Foo_sub61 struct {
	Cliui                 string `json:"cliui"`
	Decamelize            string `json:"decamelize"`
	Find_up               string `json:"find-up"`
	Get_caller_file       string `json:"get-caller-file"`
	Require_directory     string `json:"require-directory"`
	Require_main_filename string `json:"require-main-filename"`
	Set_blocking          string `json:"set-blocking"`
	String_width          string `json:"string-width"`
	Which_module          string `json:"which-module"`
	Y18n                  string `json:"y18n"`
	Yargs_parser          string `json:"yargs-parser"`
}

type Foo_sub68 struct {
	Codecov                   string `json:"codecov"`
	Cz_conventional_changelog string `json:"cz-conventional-changelog"`
	Git_cz                    string `json:"git-cz"`
	Kcd_scripts               string `json:"kcd-scripts"`
	Nock                      string `json:"nock"`
	Semantic_release          string `json:"semantic-release"`
}

type Foo_sub42 struct {
	CollectCoverage        bool      `json:"collectCoverage"`
	CollectCoverageFrom    []string  `json:"collectCoverageFrom"`
	CoverageThreshold      Foo_sub41 `json:"coverageThreshold"`
	Notify                 bool      `json:"notify"`
	TestEnvironment        string    `json:"testEnvironment"`
	TestPathIgnorePatterns []string  `json:"testPathIgnorePatterns"`
	Verbose                bool      `json:"verbose"`
}

type Foo_sub95 struct {
	Color_convert Foo_sub94 `json:"color-convert"`
}

type Foo_sub88 struct {
	Color_convert string `json:"color-convert"`
}

type Foo_sub136 struct {
	Color_name Foo_sub135 `json:"color-name"`
}

type Foo_sub89 struct {
	Color_name string `json:"color-name"`
}

type Foo_sub38 struct {
	Commit_msg string `json:"commit-msg"`
	Post_merge string `json:"post-merge"`
	Pre_commit string `json:"pre-commit"`
	Pre_push   string `json:"pre-push"`
}

type Foo_sub16 struct {
	Commitizen Foo_sub15 `json:"commitizen"`
}

type Foo_sub37 struct {
	Commitlint_cli                                string `json:"@commitlint/cli"`
	Commitlint_config_conventional                string `json:"@commitlint/config-conventional"`
	Semantic_release_changelog                    string `json:"@semantic-release/changelog"`
	Semantic_release_git                          string `json:"@semantic-release/git"`
	Semantic_release_npm                          string `json:"@semantic-release/npm"`
	Codecov                                       string `json:"codecov"`
	Cz_conventional_changelog                     string `json:"cz-conventional-changelog"`
	Eslint                                        string `json:"eslint"`
	Eslint_plugin_jest                            string `json:"eslint-plugin-jest"`
	Eslint_plugin_jquery                          string `json:"eslint-plugin-jquery"`
	Eslint_plugin_node                            string `json:"eslint-plugin-node"`
	Eslint_plugin_security                        string `json:"eslint-plugin-security"`
	Eslint_plugin_standard                        string `json:"eslint-plugin-standard"`
	Eslint_plugin_you_dont_need_lodash_underscore string `json:"eslint-plugin-you-dont-need-lodash-underscore"`
	Git_cz                                        string `json:"git-cz"`
	Husky                                         string `json:"husky"`
	Jest                                          string `json:"jest"`
	Lint_staged                                   string `json:"lint-staged"`
	Lockfile_lint                                 string `json:"lockfile-lint"`
	Prettier                                      string `json:"prettier"`
	Remark_cli                                    string `json:"remark-cli"`
	Remark_preset_lint_consistent                 string `json:"remark-preset-lint-consistent"`
	Remark_preset_lint_recommended                string `json:"remark-preset-lint-recommended"`
	Semantic_release                              string `json:"semantic-release"`
	Snyk                                          string `json:"snyk"`
}

type Foo_sub161 struct {
	Compile string `json:"compile"`
	Lint    string `json:"lint"`
	Test    string `json:"test"`
}

type Foo_sub153 struct {
	Coverage      string `json:"coverage"`
	Coverage_open string `json:"coverage-open"`
	Test          string `json:"test"`
}

type Foo_sub118 struct {
	Coverage  string `json:"coverage"`
	Coveralls string `json:"coveralls"`
	Release   string `json:"release"`
	Test      string `json:"test"`
}

type Foo_sub311 struct {
	Coverage string `json:"coverage"`
	Fix      string `json:"fix"`
	Posttest string `json:"posttest"`
	Test     string `json:"test"`
}

type Foo_sub20 struct {
	Coverage           string `json:"coverage"`
	Jsdoc              string `json:"jsdoc"`
	Lint               string `json:"lint"`
	Mocha_browser_test string `json:"mocha-browser-test"`
	Mocha_node_test    string `json:"mocha-node-test"`
	Mocha_test         string `json:"mocha-test"`
	Test               string `json:"test"`
}

type Foo_sub295 struct {
	Coverage string `json:"coverage"`
	Pretest  string `json:"pretest"`
	Release  string `json:"release"`
	Test     string `json:"test"`
}

type Foo_sub261 struct {
	Coverage string `json:"coverage"`
	Pretest  string `json:"pretest"`
	Test     string `json:"test"`
	Version  string `json:"version"`
}

type Foo_sub132 struct {
	Coverage string `json:"coverage"`
	Pretest  string `json:"pretest"`
	Test     string `json:"test"`
}

type Foo_sub117 struct {
	Coveralls        string `json:"coveralls"`
	Nyc              string `json:"nyc"`
	Standard_version string `json:"standard-version"`
	Tap_spec         string `json:"tap-spec"`
	Tape             string `json:"tape"`
}

type Foo_sub121 struct {
	Data_cover_never []string `json:"data-cover-never"`
	Output_reporter  string   `json:"output-reporter"`
	Pattern          []string `json:"pattern"`
}

type Foo_sub254 struct {
	Default string `json:"default"`
	Import  string `json:"import"`
	Module  string `json:"module"`
}

type Foo_sub7 struct {
	Default string `json:"default"`
	Import  string `json:"import"`
	Node    string `json:"node"`
}

type Foo_sub92 struct {
	Default_case       int64 `json:"default-case"`
	No_inline_comments int64 `json:"no-inline-comments"`
	Operator_linebreak int64 `json:"operator-linebreak"`
}

type Foo_sub141 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       Foo_sub3  `json:"repository"`
	Version          string    `json:"version"`
}

type Foo_sub29 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       Foo_sub3  `json:"repository"`
	Version          string    `json:"version"`
}

type Foo_sub259 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub179 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub180 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub181 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub179 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub180 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub145 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub139 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Maintainers      []string   `json:"maintainers"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub164 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub139 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Maintainers      []string   `json:"maintainers"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub140 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub139 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub274 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub139 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub207 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub206 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub265 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub206 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub237 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub56 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub57 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub58 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub56 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub57 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub281 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub76 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Types            string    `json:"types"`
	TypesVersions    Foo_sub78 `json:"typesVersions"`
	Version          string    `json:"version"`
	Xo               Foo_sub80 `json:"xo"`
}

type Foo_sub81 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub76 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Types            string    `json:"types"`
	TypesVersions    Foo_sub78 `json:"typesVersions"`
	Version          string    `json:"version"`
	Xo               Foo_sub80 `json:"xo"`
}

type Foo_sub100 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub85 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub87 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub85 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub86 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub172 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub85 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub268 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub85 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub86 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub144 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub62  `json:"author"`
	Bugs             string     `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub142 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub143 `json:"scripts"`
	Types            string     `json:"types"`
	Version          string     `json:"version"`
}

type Foo_sub264 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub62  `json:"author"`
	Bugs             string     `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub142 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub143 `json:"scripts"`
	Types            string     `json:"types"`
	Version          string     `json:"version"`
}

type Foo_sub236 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bin              Foo_sub208 `json:"bin"`
	Bugs             string     `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub209 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub210 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bin              Foo_sub208 `json:"bin"`
	Bugs             string     `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub209 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub288 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bin              Foo_sub282 `json:"bin"`
	Browser          Foo_sub283 `json:"browser"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub284 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Overrides        Foo_sub285 `json:"overrides"`
	PackageManager   string     `json:"packageManager"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub286 `json:"scripts"`
	Typings          string     `json:"typings"`
	Version          string     `json:"version"`
	Volta            Foo_sub287 `json:"volta"`
}

type Foo_sub238 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bin              Foo_sub59 `json:"bin"`
	Browser          string    `json:"browser"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Homepage         string    `json:"homepage"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Unpkg            string    `json:"unpkg"`
	Version          string    `json:"version"`
}

type Foo_sub60 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bin              Foo_sub59 `json:"bin"`
	Browser          string    `json:"browser"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Homepage         string    `json:"homepage"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Unpkg            string    `json:"unpkg"`
	Version          string    `json:"version"`
}

type Foo_sub110 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub107 `json:"bugs"`
	Contributors     []string   `json:"contributors"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub108 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Engine           Foo_sub6   `json:"engine"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	ReadmeFilename   string     `json:"readmeFilename"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub109 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub149 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub107 `json:"bugs"`
	Contributors     []string   `json:"contributors"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub108 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Engine           Foo_sub6   `json:"engine"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	ReadmeFilename   string     `json:"readmeFilename"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub109 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub243 struct {
	Dependencies     Foo_sub1     `json:"_dependencies"`
	ID               string       `json:"_id"`
	Author           string       `json:"author"`
	Bugs             Foo_sub14    `json:"bugs"`
	Contributors     []Foo_sub240 `json:"contributors"`
	Description      string       `json:"description"`
	DevDependencies  Foo_sub241   `json:"devDependencies"`
	Engines          Foo_sub6     `json:"engines"`
	Extraneous       bool         `json:"extraneous"`
	Homepage         string       `json:"homepage"`
	Keywords         []string     `json:"keywords"`
	License          string       `json:"license"`
	Main             string       `json:"main"`
	Name             string       `json:"name"`
	Overridden       bool         `json:"overridden"`
	Path             string       `json:"path"`
	PeerDependencies Foo_sub1     `json:"peerDependencies"`
	Problems         []string     `json:"problems"`
	Repository       Foo_sub3     `json:"repository"`
	Scripts          Foo_sub242   `json:"scripts"`
	Version          string       `json:"version"`
}

type Foo_sub304 struct {
	Dependencies     Foo_sub1     `json:"_dependencies"`
	ID               string       `json:"_id"`
	Author           string       `json:"author"`
	Bugs             Foo_sub14    `json:"bugs"`
	Contributors     []Foo_sub240 `json:"contributors"`
	Description      string       `json:"description"`
	DevDependencies  Foo_sub241   `json:"devDependencies"`
	Engines          Foo_sub6     `json:"engines"`
	Extraneous       bool         `json:"extraneous"`
	Homepage         string       `json:"homepage"`
	Keywords         []string     `json:"keywords"`
	License          string       `json:"license"`
	Main             string       `json:"main"`
	Name             string       `json:"name"`
	Path             string       `json:"path"`
	PeerDependencies Foo_sub1     `json:"peerDependencies"`
	Repository       Foo_sub3     `json:"repository"`
	Scripts          Foo_sub242   `json:"scripts"`
	Version          string       `json:"version"`
}

type Foo_sub280 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub1   `json:"devDependencies"`
	Exports          Foo_sub255 `json:"exports"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Jsnext_main      string     `json:"jsnext:main"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Module           string     `json:"module"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	SideEffects      bool       `json:"sideEffects"`
	Typings          string     `json:"typings"`
	Version          string     `json:"version"`
}

type Foo_sub256 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub1   `json:"devDependencies"`
	Exports          Foo_sub255 `json:"exports"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Jsnext_main      string     `json:"jsnext:main"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Module           string     `json:"module"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	SideEffects      bool       `json:"sideEffects"`
	Typings          string     `json:"typings"`
	Version          string     `json:"version"`
}

type Foo_sub138 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       Foo_sub3  `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub135 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       Foo_sub3  `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub119 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub117 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub118 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub186 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub117 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub118 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub178 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub176 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub177 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub303 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub176 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub177 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub99 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub18 `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Module           string    `json:"module"`
	Name             string    `json:"name"`
	Nyc              Foo_sub19 `json:"nyc"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       Foo_sub3  `json:"repository"`
	Scripts          Foo_sub20 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub21 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub18 `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Module           string    `json:"module"`
	Name             string    `json:"name"`
	Nyc              Foo_sub19 `json:"nyc"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       Foo_sub3  `json:"repository"`
	Scripts          Foo_sub20 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub246 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub244 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub245 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub305 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub244 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub245 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub263 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub248 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub249 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub250 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub248 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub249 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub262 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub260 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub261 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub306 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub260 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub261 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub279 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub277 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub278 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub290 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub277 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub278 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub296 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub294 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub295 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub307 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub294 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub295 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub300 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub299 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub295 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub308 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub299 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub295 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub218 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Contributors     []string  `json:"contributors"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Icon             string    `json:"icon"`
	Keywords         string    `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub47 struct {
	Dependencies     Foo_sub1  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Contributors     []string  `json:"contributors"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	Icon             string    `json:"icon"`
	Keywords         string    `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub239 struct {
	Dependencies     Foo_sub1 `json:"_dependencies"`
	ID               string   `json:"_id"`
	Author           string   `json:"author"`
	Description      string   `json:"description"`
	DevDependencies  Foo_sub1 `json:"devDependencies"`
	Extraneous       bool     `json:"extraneous"`
	Keywords         []string `json:"keywords"`
	License          string   `json:"license"`
	Main             string   `json:"main"`
	Name             string   `json:"name"`
	Overridden       bool     `json:"overridden"`
	Path             string   `json:"path"`
	PeerDependencies Foo_sub1 `json:"peerDependencies"`
	Problems         []string `json:"problems"`
	Repository       Foo_sub3 `json:"repository"`
	SideEffects      bool     `json:"sideEffects"`
	Version          string   `json:"version"`
}

type Foo_sub4 struct {
	Dependencies     Foo_sub1 `json:"_dependencies"`
	ID               string   `json:"_id"`
	Author           string   `json:"author"`
	Description      string   `json:"description"`
	DevDependencies  Foo_sub1 `json:"devDependencies"`
	Extraneous       bool     `json:"extraneous"`
	Keywords         []string `json:"keywords"`
	License          string   `json:"license"`
	Main             string   `json:"main"`
	Name             string   `json:"name"`
	Path             string   `json:"path"`
	PeerDependencies Foo_sub1 `json:"peerDependencies"`
	Repository       Foo_sub3 `json:"repository"`
	SideEffects      bool     `json:"sideEffects"`
	Version          string   `json:"version"`
}

type Foo_sub219 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub189 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub190 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub189 `json:"devDependencies"`
	Directories      Foo_sub46  `json:"directories"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub253 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub191 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub192 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub191 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub289 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub191 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub291 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub191 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub273 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub201 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Testling         Foo_sub202 `json:"testling"`
	Version          string     `json:"version"`
}

type Foo_sub203 struct {
	Dependencies     Foo_sub1   `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub201 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Testling         Foo_sub202 `json:"testling"`
	Version          string     `json:"version"`
}

type Foo_sub272 struct {
	Dependencies     Foo_sub102 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Browser          string     `json:"browser"`
	Dependencies     Foo_sub271 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub103 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub104 struct {
	Dependencies     Foo_sub102 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Browser          string     `json:"browser"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub103 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub116 struct {
	Dependencies     Foo_sub111 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub115 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub83  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub185 struct {
	Dependencies     Foo_sub111 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub83  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub252 struct {
	Dependencies     Foo_sub112 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub251 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub113 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub114 struct {
	Dependencies     Foo_sub112 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub113 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub74 struct {
	Dependencies         Foo_sub12 `json:"_dependencies"`
	ID                   string    `json:"_id"`
	Author               string    `json:"author"`
	Bin                  Foo_sub13 `json:"bin"`
	Bugs                 Foo_sub14 `json:"bugs"`
	Config               Foo_sub16 `json:"config"`
	Dependencies         Foo_sub67 `json:"dependencies"`
	Description          string    `json:"description"`
	DevDependencies      Foo_sub68 `json:"devDependencies"`
	Engines              Foo_sub6  `json:"engines"`
	EslintConfig         Foo_sub70 `json:"eslintConfig"`
	EslintIgnore         []string  `json:"eslintIgnore"`
	Extraneous           bool      `json:"extraneous"`
	Files                []string  `json:"files"`
	Homepage             string    `json:"homepage"`
	Husky                Foo_sub72 `json:"husky"`
	Keywords             []string  `json:"keywords"`
	License              string    `json:"license"`
	Main                 string    `json:"main"`
	Name                 string    `json:"name"`
	OptionalDependencies Foo_sub59 `json:"optionalDependencies"`
	Overridden           bool      `json:"overridden"`
	Path                 string    `json:"path"`
	PeerDependencies     Foo_sub1  `json:"peerDependencies"`
	Problems             []string  `json:"problems"`
	Repository           Foo_sub3  `json:"repository"`
	Scripts              Foo_sub73 `json:"scripts"`
	Version              string    `json:"version"`
}

type Foo_sub134 struct {
	Dependencies     Foo_sub120 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Config           Foo_sub122 `json:"config"`
	Dependencies     Foo_sub130 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub131 `json:"devDependencies"`
	Engine           Foo_sub6   `json:"engine"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub132 `json:"scripts"`
	Standard         Foo_sub133 `json:"standard"`
	Version          string     `json:"version"`
}

type Foo_sub301 struct {
	Dependencies     Foo_sub120 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Config           Foo_sub122 `json:"config"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub131 `json:"devDependencies"`
	Engine           Foo_sub6   `json:"engine"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub132 `json:"scripts"`
	Standard         Foo_sub133 `json:"standard"`
	Version          string     `json:"version"`
}

type Foo_sub267 struct {
	Dependencies     Foo_sub123 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub266 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub124 struct {
	Dependencies     Foo_sub123 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub270 struct {
	Dependencies     Foo_sub125 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub269 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub126 struct {
	Dependencies     Foo_sub125 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub298 struct {
	Dependencies     Foo_sub127 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub297 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub128 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub129 struct {
	Dependencies     Foo_sub127 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub128 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub162 struct {
	Dependencies     Foo_sub146 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Config           Foo_sub148 `json:"config"`
	Dependencies     Foo_sub159 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub160 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub161 `json:"scripts"`
	Types            string     `json:"types"`
	Version          string     `json:"version"`
}

type Foo_sub187 struct {
	Dependencies     Foo_sub146 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Config           Foo_sub148 `json:"config"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub160 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub161 `json:"scripts"`
	Types            string     `json:"types"`
	Version          string     `json:"version"`
}

type Foo_sub183 struct {
	Dependencies     Foo_sub150 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Browser          Foo_sub151 `json:"browser"`
	Bugs             string     `json:"bugs"`
	Dependencies     Foo_sub182 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub152 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub153 `json:"scripts"`
	Typings          string     `json:"typings"`
	Version          string     `json:"version"`
}

type Foo_sub154 struct {
	Dependencies     Foo_sub150 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Browser          Foo_sub151 `json:"browser"`
	Bugs             string     `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub152 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub153 `json:"scripts"`
	Typings          string     `json:"typings"`
	Version          string     `json:"version"`
}

type Foo_sub276 struct {
	Dependencies     Foo_sub155 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Dependencies     Foo_sub275 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub156 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub157 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub158 struct {
	Dependencies     Foo_sub155 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub156 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub157 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub168 struct {
	Dependencies     Foo_sub163 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub165 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub166 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub167 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub188 struct {
	Dependencies     Foo_sub163 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub166 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub167 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub175 struct {
	Dependencies     Foo_sub169 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub173 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub174 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub302 struct {
	Dependencies     Foo_sub169 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub174 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub217 struct {
	Dependencies     Foo_sub170 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub216 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub171 struct {
	Dependencies     Foo_sub170 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub258 struct {
	Dependencies     Foo_sub193  `json:"_dependencies"`
	ID               string      `json:"_id"`
	Author           string      `json:"author"`
	Bugs             Foo_sub14   `json:"bugs"`
	Config           Foo_sub16   `json:"config"`
	Contributors     []Foo_sub35 `json:"contributors"`
	Dependencies     Foo_sub257  `json:"dependencies"`
	Description      string      `json:"description"`
	DevDependencies  Foo_sub194  `json:"devDependencies"`
	Engines          Foo_sub195  `json:"engines"`
	Es2015           string      `json:"es2015"`
	Extraneous       bool        `json:"extraneous"`
	Homepage         string      `json:"homepage"`
	Keywords         []string    `json:"keywords"`
	License          string      `json:"license"`
	Lint_staged      Foo_sub197  `json:"lint-staged"`
	Main             string      `json:"main"`
	Module           string      `json:"module"`
	Name             string      `json:"name"`
	Ng_update        Foo_sub198  `json:"ng-update"`
	Nyc              Foo_sub199  `json:"nyc"`
	Overridden       bool        `json:"overridden"`
	Path             string      `json:"path"`
	PeerDependencies Foo_sub1    `json:"peerDependencies"`
	Problems         []string    `json:"problems"`
	Repository       Foo_sub3    `json:"repository"`
	SideEffects      bool        `json:"sideEffects"`
	Typings          string      `json:"typings"`
	Version          string      `json:"version"`
}

type Foo_sub200 struct {
	Dependencies     Foo_sub193  `json:"_dependencies"`
	ID               string      `json:"_id"`
	Author           string      `json:"author"`
	Bugs             Foo_sub14   `json:"bugs"`
	Config           Foo_sub16   `json:"config"`
	Contributors     []Foo_sub35 `json:"contributors"`
	Description      string      `json:"description"`
	DevDependencies  Foo_sub194  `json:"devDependencies"`
	Engines          Foo_sub195  `json:"engines"`
	Es2015           string      `json:"es2015"`
	Extraneous       bool        `json:"extraneous"`
	Homepage         string      `json:"homepage"`
	Keywords         []string    `json:"keywords"`
	License          string      `json:"license"`
	Lint_staged      Foo_sub197  `json:"lint-staged"`
	Main             string      `json:"main"`
	Module           string      `json:"module"`
	Name             string      `json:"name"`
	Ng_update        Foo_sub198  `json:"ng-update"`
	Nyc              Foo_sub199  `json:"nyc"`
	Path             string      `json:"path"`
	PeerDependencies Foo_sub1    `json:"peerDependencies"`
	Repository       Foo_sub3    `json:"repository"`
	SideEffects      bool        `json:"sideEffects"`
	Typings          string      `json:"typings"`
	Version          string      `json:"version"`
}

type Foo_sub11 struct {
	Dependencies     Foo_sub2  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Dependencies     Foo_sub5  `json:"dependencies"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Exports          Foo_sub8  `json:"exports"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	PublishConfig    Foo_sub9  `json:"publishConfig"`
	Repository       Foo_sub10 `json:"repository"`
	Type             string    `json:"type"`
	Version          string    `json:"version"`
}

type Foo_sub17 struct {
	Dependencies     Foo_sub2  `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub1  `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Exports          Foo_sub8  `json:"exports"`
	Extraneous       bool      `json:"extraneous"`
	Homepage         string    `json:"homepage"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	PublishConfig    Foo_sub9  `json:"publishConfig"`
	Repository       Foo_sub10 `json:"repository"`
	Type             string    `json:"type"`
	Version          string    `json:"version"`
}

type Foo_sub235 struct {
	Dependencies     Foo_sub213 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub234 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub214 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub215 struct {
	Dependencies     Foo_sub213 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub214 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub106 struct {
	Dependencies     Foo_sub22  `json:"_dependencies"`
	ID               string     `json:"_id"`
	Dependencies     Foo_sub105 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub23  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub24  `json:"scripts"`
	Version          string     `json:"version"`
	Xo               Foo_sub26  `json:"xo"`
}

type Foo_sub27 struct {
	Dependencies     Foo_sub22 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub23 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub24 `json:"scripts"`
	Version          string    `json:"version"`
	Xo               Foo_sub26 `json:"xo"`
}

type Foo_sub293 struct {
	Dependencies     Foo_sub220 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Dependencies     Foo_sub292 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub221 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub222 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub223 struct {
	Dependencies     Foo_sub220 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub221 `json:"devDependencies"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub222 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub228 struct {
	Dependencies     Foo_sub226 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub227 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub247 struct {
	Dependencies     Foo_sub226 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub85  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub232 struct {
	Dependencies     Foo_sub229 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Dependencies     Foo_sub230 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub231 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub233 struct {
	Dependencies     Foo_sub229 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub28  `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub231 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Funding          Foo_sub14  `json:"funding"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub46  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub205 struct {
	Dependencies     Foo_sub30  `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Dependencies     Foo_sub204 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub31  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	GitHead          string     `json:"gitHead"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub32  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub33 struct {
	Dependencies     Foo_sub30 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub31 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	GitHead          string    `json:"gitHead"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub32 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub316 struct {
	Dependencies     Foo_sub309 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Dependencies     Foo_sub315 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub310 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub311 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub312 struct {
	Dependencies     Foo_sub309 `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub310 `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Repository       Foo_sub3   `json:"repository"`
	Scripts          Foo_sub311 `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub212 struct {
	Dependencies     Foo_sub34  `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           Foo_sub35  `json:"author"`
	Bugs             Foo_sub14  `json:"bugs"`
	Commitlint       Foo_sub36  `json:"commitlint"`
	Config           Foo_sub16  `json:"config"`
	Dependencies     Foo_sub211 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub37  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Homepage         string     `json:"homepage"`
	Husky            Foo_sub39  `json:"husky"`
	Jest             Foo_sub42  `json:"jest"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Lint_staged      Foo_sub43  `json:"lint-staged"`
	Main             string     `json:"main"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Private          bool       `json:"private"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub44  `json:"scripts"`
	Version          string     `json:"version"`
}

type Foo_sub45 struct {
	Dependencies     Foo_sub34 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub35 `json:"author"`
	Bugs             Foo_sub14 `json:"bugs"`
	Commitlint       Foo_sub36 `json:"commitlint"`
	Config           Foo_sub16 `json:"config"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub37 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Homepage         string    `json:"homepage"`
	Husky            Foo_sub39 `json:"husky"`
	Jest             Foo_sub42 `json:"jest"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Lint_staged      Foo_sub43 `json:"lint-staged"`
	Main             string    `json:"main"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Private          bool      `json:"private"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub44 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub225 struct {
	Dependencies         Foo_sub48  `json:"_dependencies"`
	ID                   string     `json:"_id"`
	Author               string     `json:"author"`
	Browser              string     `json:"browser"`
	Bugs                 Foo_sub14  `json:"bugs"`
	Dependencies         Foo_sub224 `json:"dependencies"`
	Description          string     `json:"description"`
	DevDependencies      Foo_sub49  `json:"devDependencies"`
	Engines              Foo_sub6   `json:"engines"`
	Extraneous           bool       `json:"extraneous"`
	Files                []string   `json:"files"`
	Homepage             string     `json:"homepage"`
	Keywords             []string   `json:"keywords"`
	License              string     `json:"license"`
	Main                 string     `json:"main"`
	Module               string     `json:"module"`
	Name                 string     `json:"name"`
	Overridden           bool       `json:"overridden"`
	Path                 string     `json:"path"`
	PeerDependencies     Foo_sub50  `json:"peerDependencies"`
	PeerDependenciesMeta Foo_sub52  `json:"peerDependenciesMeta"`
	Problems             []string   `json:"problems"`
	Release              Foo_sub53  `json:"release"`
	Repository           Foo_sub3   `json:"repository"`
	Scripts              Foo_sub54  `json:"scripts"`
	Version              string     `json:"version"`
}

type Foo_sub55 struct {
	Dependencies         Foo_sub48 `json:"_dependencies"`
	ID                   string    `json:"_id"`
	Author               string    `json:"author"`
	Browser              string    `json:"browser"`
	Bugs                 Foo_sub14 `json:"bugs"`
	Description          string    `json:"description"`
	DevDependencies      Foo_sub49 `json:"devDependencies"`
	Engines              Foo_sub6  `json:"engines"`
	Extraneous           bool      `json:"extraneous"`
	Files                []string  `json:"files"`
	Homepage             string    `json:"homepage"`
	Keywords             []string  `json:"keywords"`
	License              string    `json:"license"`
	Main                 string    `json:"main"`
	Module               string    `json:"module"`
	Name                 string    `json:"name"`
	Path                 string    `json:"path"`
	PeerDependencies     Foo_sub50 `json:"peerDependencies"`
	PeerDependenciesMeta Foo_sub52 `json:"peerDependenciesMeta"`
	Release              Foo_sub53 `json:"release"`
	Repository           Foo_sub3  `json:"repository"`
	Scripts              Foo_sub54 `json:"scripts"`
	Version              string    `json:"version"`
}

type Foo_sub314 struct {
	Dependencies     Foo_sub61   `json:"_dependencies"`
	ID               string      `json:"_id"`
	Contributors     []Foo_sub62 `json:"contributors"`
	Dependencies     Foo_sub313  `json:"dependencies"`
	Description      string      `json:"description"`
	DevDependencies  Foo_sub63   `json:"devDependencies"`
	Engines          Foo_sub6    `json:"engines"`
	Extraneous       bool        `json:"extraneous"`
	Files            []string    `json:"files"`
	Homepage         string      `json:"homepage"`
	Keywords         []string    `json:"keywords"`
	License          string      `json:"license"`
	Main             string      `json:"main"`
	Name             string      `json:"name"`
	Overridden       bool        `json:"overridden"`
	Path             string      `json:"path"`
	PeerDependencies Foo_sub1    `json:"peerDependencies"`
	Problems         []string    `json:"problems"`
	Repository       Foo_sub3    `json:"repository"`
	Scripts          Foo_sub64   `json:"scripts"`
	Standardx        Foo_sub65   `json:"standardx"`
	Version          string      `json:"version"`
}

type Foo_sub66 struct {
	Dependencies     Foo_sub61   `json:"_dependencies"`
	ID               string      `json:"_id"`
	Contributors     []Foo_sub62 `json:"contributors"`
	Description      string      `json:"description"`
	DevDependencies  Foo_sub63   `json:"devDependencies"`
	Engines          Foo_sub6    `json:"engines"`
	Extraneous       bool        `json:"extraneous"`
	Files            []string    `json:"files"`
	Homepage         string      `json:"homepage"`
	Keywords         []string    `json:"keywords"`
	License          string      `json:"license"`
	Main             string      `json:"main"`
	Name             string      `json:"name"`
	Path             string      `json:"path"`
	PeerDependencies Foo_sub1    `json:"peerDependencies"`
	Repository       Foo_sub3    `json:"repository"`
	Scripts          Foo_sub64   `json:"scripts"`
	Standardx        Foo_sub65   `json:"standardx"`
	Version          string      `json:"version"`
}

type Foo_sub84 struct {
	Dependencies     Foo_sub75 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Dependencies     Foo_sub82 `json:"dependencies"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub83 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub184 struct {
	Dependencies     Foo_sub75 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub83 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub46 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub98 struct {
	Dependencies     Foo_sub88 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Dependencies     Foo_sub95 `json:"dependencies"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub96 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Overridden       bool      `json:"overridden"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Problems         []string  `json:"problems"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub97 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub101 struct {
	Dependencies     Foo_sub88 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           Foo_sub28 `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub96 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Funding          Foo_sub14 `json:"funding"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub97 `json:"scripts"`
	Version          string    `json:"version"`
}

type Foo_sub137 struct {
	Dependencies     Foo_sub89  `json:"_dependencies"`
	ID               string     `json:"_id"`
	Author           string     `json:"author"`
	Dependencies     Foo_sub136 `json:"dependencies"`
	Description      string     `json:"description"`
	DevDependencies  Foo_sub90  `json:"devDependencies"`
	Engines          Foo_sub6   `json:"engines"`
	Extraneous       bool       `json:"extraneous"`
	Files            []string   `json:"files"`
	Keywords         []string   `json:"keywords"`
	License          string     `json:"license"`
	Name             string     `json:"name"`
	Overridden       bool       `json:"overridden"`
	Path             string     `json:"path"`
	PeerDependencies Foo_sub1   `json:"peerDependencies"`
	Problems         []string   `json:"problems"`
	Repository       string     `json:"repository"`
	Scripts          Foo_sub91  `json:"scripts"`
	Version          string     `json:"version"`
	Xo               Foo_sub93  `json:"xo"`
}

type Foo_sub94 struct {
	Dependencies     Foo_sub89 `json:"_dependencies"`
	ID               string    `json:"_id"`
	Author           string    `json:"author"`
	Description      string    `json:"description"`
	DevDependencies  Foo_sub90 `json:"devDependencies"`
	Engines          Foo_sub6  `json:"engines"`
	Extraneous       bool      `json:"extraneous"`
	Files            []string  `json:"files"`
	Keywords         []string  `json:"keywords"`
	License          string    `json:"license"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	PeerDependencies Foo_sub1  `json:"peerDependencies"`
	Repository       string    `json:"repository"`
	Scripts          Foo_sub91 `json:"scripts"`
	Version          string    `json:"version"`
	Xo               Foo_sub93 `json:"xo"`
}

type Foo_sub10 struct {
	Directory string `json:"directory"`
	Type      string `json:"type"`
	URL       string `json:"url"`
}

type Foo_sub157 struct {
	Doc  string `json:"doc"`
	Test string `json:"test"`
}

type Foo_sub28 struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

type Foo_sub240 struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Web   string `json:"web"`
}

type Foo_sub35 struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Foo_sub266 struct {
	Emoji_regex             Foo_sub264 `json:"emoji-regex"`
	Is_fullwidth_code_point Foo_sub265 `json:"is-fullwidth-code-point"`
	Strip_ansi              Foo_sub126 `json:"strip-ansi"`
}

type Foo_sub123 struct {
	Emoji_regex             string `json:"emoji-regex"`
	Is_fullwidth_code_point string `json:"is-fullwidth-code-point"`
	Strip_ansi              string `json:"strip-ansi"`
}

type Foo_sub224 struct {
	Encoding   Foo_sub1   `json:"encoding"`
	Whatwg_url Foo_sub223 `json:"whatwg-url"`
}

type Foo_sub52 struct {
	Encoding Foo_sub51 `json:"encoding"`
}

type Foo_sub50 struct {
	Encoding string `json:"encoding"`
}

type Foo_sub165 struct {
	Escape_string_regexp Foo_sub164 `json:"escape-string-regexp"`
}

type Foo_sub163 struct {
	Escape_string_regexp string `json:"escape-string-regexp"`
}

type Foo_sub284 struct {
	Esfx_canceltoken                      string `json:"@esfx/canceltoken"`
	Octokit_rest                          string `json:"@octokit/rest"`
	Types_chai                            string `json:"@types/chai"`
	Types_fs_extra                        string `json:"@types/fs-extra"`
	Types_glob                            string `json:"@types/glob"`
	Types_microsoftTypescript_etw         string `json:"@types/microsoft__typescript-etw"`
	Types_minimist                        string `json:"@types/minimist"`
	Types_mocha                           string `json:"@types/mocha"`
	Types_ms                              string `json:"@types/ms"`
	Types_node                            string `json:"@types/node"`
	Types_source_map_support              string `json:"@types/source-map-support"`
	Types_which                           string `json:"@types/which"`
	Typescript_eslint_eslint_plugin       string `json:"@typescript-eslint/eslint-plugin"`
	Typescript_eslint_parser              string `json:"@typescript-eslint/parser"`
	Typescript_eslint_utils               string `json:"@typescript-eslint/utils"`
	Azure_devops_node_api                 string `json:"azure-devops-node-api"`
	Chai                                  string `json:"chai"`
	Chalk                                 string `json:"chalk"`
	Chokidar                              string `json:"chokidar"`
	Del                                   string `json:"del"`
	Diff                                  string `json:"diff"`
	Esbuild                               string `json:"esbuild"`
	Eslint                                string `json:"eslint"`
	Eslint_formatter_autolinkable_stylish string `json:"eslint-formatter-autolinkable-stylish"`
	Eslint_plugin_import                  string `json:"eslint-plugin-import"`
	Eslint_plugin_local                   string `json:"eslint-plugin-local"`
	Eslint_plugin_no_null                 string `json:"eslint-plugin-no-null"`
	Eslint_plugin_simple_import_sort      string `json:"eslint-plugin-simple-import-sort"`
	Fast_xml_parser                       string `json:"fast-xml-parser"`
	Fs_extra                              string `json:"fs-extra"`
	Glob                                  string `json:"glob"`
	Hereby                                string `json:"hereby"`
	Jsonc_parser                          string `json:"jsonc-parser"`
	Minimist                              string `json:"minimist"`
	Mocha                                 string `json:"mocha"`
	Mocha_fivemat_progress_reporter       string `json:"mocha-fivemat-progress-reporter"`
	Ms                                    string `json:"ms"`
	Node_fetch                            string `json:"node-fetch"`
	Source_map_support                    string `json:"source-map-support"`
	Tslib                                 string `json:"tslib"`
	Typescript                            string `json:"typescript"`
	Which                                 string `json:"which"`
}

type Foo_sub221 struct {
	Eslint    string `json:"eslint"`
	Istanbul  string `json:"istanbul"`
	Mocha     string `json:"mocha"`
	Recast    string `json:"recast"`
	Request   string `json:"request"`
	Webidl2js string `json:"webidl2js"`
}

type Foo_sub19 struct {
	Exclude []string `json:"exclude"`
}

type Foo_sub36 struct {
	Extends []string `json:"extends"`
}

type Foo_sub70 struct {
	Extends string    `json:"extends"`
	Rules   Foo_sub69 `json:"rules"`
}

type Foo_sub78 struct {
	Four_1 Foo_sub77 `json:">=4.1"`
}

type Foo_sub201 struct {
	From        string `json:"from"`
	Stream_spec string `json:"stream-spec"`
	Tape        string `json:"tape"`
}

type Foo_sub108 struct {
	Github_publish_release string `json:"github-publish-release"`
	Mocha                  string `json:"mocha"`
}

type Foo_sub41 struct {
	Global Foo_sub40 `json:"global"`
}

type Foo_sub133 struct {
	Globals []string `json:"globals"`
	Ignore  []string `json:"ignore"`
}

type Foo_sub271 struct {
	Has_flag Foo_sub172 `json:"has-flag"`
}

type Foo_sub102 struct {
	Has_flag string `json:"has-flag"`
}

type Foo_sub8 struct {
	Helpers_AsyncGenerator                             []Foo_sub7 `json:"./helpers/AsyncGenerator"`
	Helpers_AwaitValue                                 []Foo_sub7 `json:"./helpers/AwaitValue"`
	Helpers_OverloadYield                              []Foo_sub7 `json:"./helpers/OverloadYield"`
	Helpers_applyDecoratedDescriptor                   []Foo_sub7 `json:"./helpers/applyDecoratedDescriptor"`
	Helpers_applyDecs                                  []Foo_sub7 `json:"./helpers/applyDecs"`
	Helpers_applyDecs2203                              []Foo_sub7 `json:"./helpers/applyDecs2203"`
	Helpers_applyDecs2203R                             []Foo_sub7 `json:"./helpers/applyDecs2203R"`
	Helpers_applyDecs2301                              []Foo_sub7 `json:"./helpers/applyDecs2301"`
	Helpers_applyDecs2305                              []Foo_sub7 `json:"./helpers/applyDecs2305"`
	Helpers_arrayLikeToArray                           []Foo_sub7 `json:"./helpers/arrayLikeToArray"`
	Helpers_arrayWithHoles                             []Foo_sub7 `json:"./helpers/arrayWithHoles"`
	Helpers_arrayWithoutHoles                          []Foo_sub7 `json:"./helpers/arrayWithoutHoles"`
	Helpers_assertThisInitialized                      []Foo_sub7 `json:"./helpers/assertThisInitialized"`
	Helpers_asyncGeneratorDelegate                     []Foo_sub7 `json:"./helpers/asyncGeneratorDelegate"`
	Helpers_asyncIterator                              []Foo_sub7 `json:"./helpers/asyncIterator"`
	Helpers_asyncToGenerator                           []Foo_sub7 `json:"./helpers/asyncToGenerator"`
	Helpers_awaitAsyncGenerator                        []Foo_sub7 `json:"./helpers/awaitAsyncGenerator"`
	Helpers_checkInRHS                                 []Foo_sub7 `json:"./helpers/checkInRHS"`
	Helpers_checkPrivateRedeclaration                  []Foo_sub7 `json:"./helpers/checkPrivateRedeclaration"`
	Helpers_classApplyDescriptorDestructureSet         []Foo_sub7 `json:"./helpers/classApplyDescriptorDestructureSet"`
	Helpers_classApplyDescriptorGet                    []Foo_sub7 `json:"./helpers/classApplyDescriptorGet"`
	Helpers_classApplyDescriptorSet                    []Foo_sub7 `json:"./helpers/classApplyDescriptorSet"`
	Helpers_classCallCheck                             []Foo_sub7 `json:"./helpers/classCallCheck"`
	Helpers_classCheckPrivateStaticAccess              []Foo_sub7 `json:"./helpers/classCheckPrivateStaticAccess"`
	Helpers_classCheckPrivateStaticFieldDescriptor     []Foo_sub7 `json:"./helpers/classCheckPrivateStaticFieldDescriptor"`
	Helpers_classExtractFieldDescriptor                []Foo_sub7 `json:"./helpers/classExtractFieldDescriptor"`
	Helpers_classNameTDZError                          []Foo_sub7 `json:"./helpers/classNameTDZError"`
	Helpers_classPrivateFieldDestructureSet            []Foo_sub7 `json:"./helpers/classPrivateFieldDestructureSet"`
	Helpers_classPrivateFieldGet                       []Foo_sub7 `json:"./helpers/classPrivateFieldGet"`
	Helpers_classPrivateFieldInitSpec                  []Foo_sub7 `json:"./helpers/classPrivateFieldInitSpec"`
	Helpers_classPrivateFieldLooseBase                 []Foo_sub7 `json:"./helpers/classPrivateFieldLooseBase"`
	Helpers_classPrivateFieldLooseKey                  []Foo_sub7 `json:"./helpers/classPrivateFieldLooseKey"`
	Helpers_classPrivateFieldSet                       []Foo_sub7 `json:"./helpers/classPrivateFieldSet"`
	Helpers_classPrivateMethodGet                      []Foo_sub7 `json:"./helpers/classPrivateMethodGet"`
	Helpers_classPrivateMethodInitSpec                 []Foo_sub7 `json:"./helpers/classPrivateMethodInitSpec"`
	Helpers_classPrivateMethodSet                      []Foo_sub7 `json:"./helpers/classPrivateMethodSet"`
	Helpers_classStaticPrivateFieldDestructureSet      []Foo_sub7 `json:"./helpers/classStaticPrivateFieldDestructureSet"`
	Helpers_classStaticPrivateFieldSpecGet             []Foo_sub7 `json:"./helpers/classStaticPrivateFieldSpecGet"`
	Helpers_classStaticPrivateFieldSpecSet             []Foo_sub7 `json:"./helpers/classStaticPrivateFieldSpecSet"`
	Helpers_classStaticPrivateMethodGet                []Foo_sub7 `json:"./helpers/classStaticPrivateMethodGet"`
	Helpers_classStaticPrivateMethodSet                []Foo_sub7 `json:"./helpers/classStaticPrivateMethodSet"`
	Helpers_construct                                  []Foo_sub7 `json:"./helpers/construct"`
	Helpers_createClass                                []Foo_sub7 `json:"./helpers/createClass"`
	Helpers_createForOfIteratorHelper                  []Foo_sub7 `json:"./helpers/createForOfIteratorHelper"`
	Helpers_createForOfIteratorHelperLoose             []Foo_sub7 `json:"./helpers/createForOfIteratorHelperLoose"`
	Helpers_createSuper                                []Foo_sub7 `json:"./helpers/createSuper"`
	Helpers_decorate                                   []Foo_sub7 `json:"./helpers/decorate"`
	Helpers_defaults                                   []Foo_sub7 `json:"./helpers/defaults"`
	Helpers_defineAccessor                             []Foo_sub7 `json:"./helpers/defineAccessor"`
	Helpers_defineEnumerableProperties                 []Foo_sub7 `json:"./helpers/defineEnumerableProperties"`
	Helpers_defineProperty                             []Foo_sub7 `json:"./helpers/defineProperty"`
	Helpers_dispose                                    []Foo_sub7 `json:"./helpers/dispose"`
	Helpers_esm_AsyncGenerator                         string     `json:"./helpers/esm/AsyncGenerator"`
	Helpers_esm_AwaitValue                             string     `json:"./helpers/esm/AwaitValue"`
	Helpers_esm_OverloadYield                          string     `json:"./helpers/esm/OverloadYield"`
	Helpers_esm_applyDecoratedDescriptor               string     `json:"./helpers/esm/applyDecoratedDescriptor"`
	Helpers_esm_applyDecs                              string     `json:"./helpers/esm/applyDecs"`
	Helpers_esm_applyDecs2203                          string     `json:"./helpers/esm/applyDecs2203"`
	Helpers_esm_applyDecs2203R                         string     `json:"./helpers/esm/applyDecs2203R"`
	Helpers_esm_applyDecs2301                          string     `json:"./helpers/esm/applyDecs2301"`
	Helpers_esm_applyDecs2305                          string     `json:"./helpers/esm/applyDecs2305"`
	Helpers_esm_arrayLikeToArray                       string     `json:"./helpers/esm/arrayLikeToArray"`
	Helpers_esm_arrayWithHoles                         string     `json:"./helpers/esm/arrayWithHoles"`
	Helpers_esm_arrayWithoutHoles                      string     `json:"./helpers/esm/arrayWithoutHoles"`
	Helpers_esm_assertThisInitialized                  string     `json:"./helpers/esm/assertThisInitialized"`
	Helpers_esm_asyncGeneratorDelegate                 string     `json:"./helpers/esm/asyncGeneratorDelegate"`
	Helpers_esm_asyncIterator                          string     `json:"./helpers/esm/asyncIterator"`
	Helpers_esm_asyncToGenerator                       string     `json:"./helpers/esm/asyncToGenerator"`
	Helpers_esm_awaitAsyncGenerator                    string     `json:"./helpers/esm/awaitAsyncGenerator"`
	Helpers_esm_checkInRHS                             string     `json:"./helpers/esm/checkInRHS"`
	Helpers_esm_checkPrivateRedeclaration              string     `json:"./helpers/esm/checkPrivateRedeclaration"`
	Helpers_esm_classApplyDescriptorDestructureSet     string     `json:"./helpers/esm/classApplyDescriptorDestructureSet"`
	Helpers_esm_classApplyDescriptorGet                string     `json:"./helpers/esm/classApplyDescriptorGet"`
	Helpers_esm_classApplyDescriptorSet                string     `json:"./helpers/esm/classApplyDescriptorSet"`
	Helpers_esm_classCallCheck                         string     `json:"./helpers/esm/classCallCheck"`
	Helpers_esm_classCheckPrivateStaticAccess          string     `json:"./helpers/esm/classCheckPrivateStaticAccess"`
	Helpers_esm_classCheckPrivateStaticFieldDescriptor string     `json:"./helpers/esm/classCheckPrivateStaticFieldDescriptor"`
	Helpers_esm_classExtractFieldDescriptor            string     `json:"./helpers/esm/classExtractFieldDescriptor"`
	Helpers_esm_classNameTDZError                      string     `json:"./helpers/esm/classNameTDZError"`
	Helpers_esm_classPrivateFieldDestructureSet        string     `json:"./helpers/esm/classPrivateFieldDestructureSet"`
	Helpers_esm_classPrivateFieldGet                   string     `json:"./helpers/esm/classPrivateFieldGet"`
	Helpers_esm_classPrivateFieldInitSpec              string     `json:"./helpers/esm/classPrivateFieldInitSpec"`
	Helpers_esm_classPrivateFieldLooseBase             string     `json:"./helpers/esm/classPrivateFieldLooseBase"`
	Helpers_esm_classPrivateFieldLooseKey              string     `json:"./helpers/esm/classPrivateFieldLooseKey"`
	Helpers_esm_classPrivateFieldSet                   string     `json:"./helpers/esm/classPrivateFieldSet"`
	Helpers_esm_classPrivateMethodGet                  string     `json:"./helpers/esm/classPrivateMethodGet"`
	Helpers_esm_classPrivateMethodInitSpec             string     `json:"./helpers/esm/classPrivateMethodInitSpec"`
	Helpers_esm_classPrivateMethodSet                  string     `json:"./helpers/esm/classPrivateMethodSet"`
	Helpers_esm_classStaticPrivateFieldDestructureSet  string     `json:"./helpers/esm/classStaticPrivateFieldDestructureSet"`
	Helpers_esm_classStaticPrivateFieldSpecGet         string     `json:"./helpers/esm/classStaticPrivateFieldSpecGet"`
	Helpers_esm_classStaticPrivateFieldSpecSet         string     `json:"./helpers/esm/classStaticPrivateFieldSpecSet"`
	Helpers_esm_classStaticPrivateMethodGet            string     `json:"./helpers/esm/classStaticPrivateMethodGet"`
	Helpers_esm_classStaticPrivateMethodSet            string     `json:"./helpers/esm/classStaticPrivateMethodSet"`
	Helpers_esm_construct                              string     `json:"./helpers/esm/construct"`
	Helpers_esm_createClass                            string     `json:"./helpers/esm/createClass"`
	Helpers_esm_createForOfIteratorHelper              string     `json:"./helpers/esm/createForOfIteratorHelper"`
	Helpers_esm_createForOfIteratorHelperLoose         string     `json:"./helpers/esm/createForOfIteratorHelperLoose"`
	Helpers_esm_createSuper                            string     `json:"./helpers/esm/createSuper"`
	Helpers_esm_decorate                               string     `json:"./helpers/esm/decorate"`
	Helpers_esm_defaults                               string     `json:"./helpers/esm/defaults"`
	Helpers_esm_defineAccessor                         string     `json:"./helpers/esm/defineAccessor"`
	Helpers_esm_defineEnumerableProperties             string     `json:"./helpers/esm/defineEnumerableProperties"`
	Helpers_esm_defineProperty                         string     `json:"./helpers/esm/defineProperty"`
	Helpers_esm_dispose                                string     `json:"./helpers/esm/dispose"`
	Helpers_esm_extends                                string     `json:"./helpers/esm/extends"`
	Helpers_esm_get                                    string     `json:"./helpers/esm/get"`
	Helpers_esm_getPrototypeOf                         string     `json:"./helpers/esm/getPrototypeOf"`
	Helpers_esm_identity                               string     `json:"./helpers/esm/identity"`
	Helpers_esm_inherits                               string     `json:"./helpers/esm/inherits"`
	Helpers_esm_inheritsLoose                          string     `json:"./helpers/esm/inheritsLoose"`
	Helpers_esm_initializerDefineProperty              string     `json:"./helpers/esm/initializerDefineProperty"`
	Helpers_esm_initializerWarningHelper               string     `json:"./helpers/esm/initializerWarningHelper"`
	Helpers_esm_instanceof                             string     `json:"./helpers/esm/instanceof"`
	Helpers_esm_interopRequireDefault                  string     `json:"./helpers/esm/interopRequireDefault"`
	Helpers_esm_interopRequireWildcard                 string     `json:"./helpers/esm/interopRequireWildcard"`
	Helpers_esm_isNativeFunction                       string     `json:"./helpers/esm/isNativeFunction"`
	Helpers_esm_isNativeReflectConstruct               string     `json:"./helpers/esm/isNativeReflectConstruct"`
	Helpers_esm_iterableToArray                        string     `json:"./helpers/esm/iterableToArray"`
	Helpers_esm_iterableToArrayLimit                   string     `json:"./helpers/esm/iterableToArrayLimit"`
	Helpers_esm_iterableToArrayLimitLoose              string     `json:"./helpers/esm/iterableToArrayLimitLoose"`
	Helpers_esm_jsx                                    string     `json:"./helpers/esm/jsx"`
	Helpers_esm_maybeArrayLike                         string     `json:"./helpers/esm/maybeArrayLike"`
	Helpers_esm_newArrowCheck                          string     `json:"./helpers/esm/newArrowCheck"`
	Helpers_esm_nonIterableRest                        string     `json:"./helpers/esm/nonIterableRest"`
	Helpers_esm_nonIterableSpread                      string     `json:"./helpers/esm/nonIterableSpread"`
	Helpers_esm_objectDestructuringEmpty               string     `json:"./helpers/esm/objectDestructuringEmpty"`
	Helpers_esm_objectSpread                           string     `json:"./helpers/esm/objectSpread"`
	Helpers_esm_objectSpread2                          string     `json:"./helpers/esm/objectSpread2"`
	Helpers_esm_objectWithoutProperties                string     `json:"./helpers/esm/objectWithoutProperties"`
	Helpers_esm_objectWithoutPropertiesLoose           string     `json:"./helpers/esm/objectWithoutPropertiesLoose"`
	Helpers_esm_possibleConstructorReturn              string     `json:"./helpers/esm/possibleConstructorReturn"`
	Helpers_esm_readOnlyError                          string     `json:"./helpers/esm/readOnlyError"`
	Helpers_esm_regeneratorRuntime                     string     `json:"./helpers/esm/regeneratorRuntime"`
	Helpers_esm_set                                    string     `json:"./helpers/esm/set"`
	Helpers_esm_setPrototypeOf                         string     `json:"./helpers/esm/setPrototypeOf"`
	Helpers_esm_skipFirstGeneratorNext                 string     `json:"./helpers/esm/skipFirstGeneratorNext"`
	Helpers_esm_slicedToArray                          string     `json:"./helpers/esm/slicedToArray"`
	Helpers_esm_slicedToArrayLoose                     string     `json:"./helpers/esm/slicedToArrayLoose"`
	Helpers_esm_superPropBase                          string     `json:"./helpers/esm/superPropBase"`
	Helpers_esm_taggedTemplateLiteral                  string     `json:"./helpers/esm/taggedTemplateLiteral"`
	Helpers_esm_taggedTemplateLiteralLoose             string     `json:"./helpers/esm/taggedTemplateLiteralLoose"`
	Helpers_esm_tdz                                    string     `json:"./helpers/esm/tdz"`
	Helpers_esm_temporalRef                            string     `json:"./helpers/esm/temporalRef"`
	Helpers_esm_temporalUndefined                      string     `json:"./helpers/esm/temporalUndefined"`
	Helpers_esm_toArray                                string     `json:"./helpers/esm/toArray"`
	Helpers_esm_toConsumableArray                      string     `json:"./helpers/esm/toConsumableArray"`
	Helpers_esm_toPrimitive                            string     `json:"./helpers/esm/toPrimitive"`
	Helpers_esm_toPropertyKey                          string     `json:"./helpers/esm/toPropertyKey"`
	Helpers_esm_typeof                                 string     `json:"./helpers/esm/typeof"`
	Helpers_esm_unsupportedIterableToArray             string     `json:"./helpers/esm/unsupportedIterableToArray"`
	Helpers_esm_using                                  string     `json:"./helpers/esm/using"`
	Helpers_esm_wrapAsyncGenerator                     string     `json:"./helpers/esm/wrapAsyncGenerator"`
	Helpers_esm_wrapNativeSuper                        string     `json:"./helpers/esm/wrapNativeSuper"`
	Helpers_esm_wrapRegExp                             string     `json:"./helpers/esm/wrapRegExp"`
	Helpers_esm_writeOnlyError                         string     `json:"./helpers/esm/writeOnlyError"`
	Helpers_extends                                    []Foo_sub7 `json:"./helpers/extends"`
	Helpers_get                                        []Foo_sub7 `json:"./helpers/get"`
	Helpers_getPrototypeOf                             []Foo_sub7 `json:"./helpers/getPrototypeOf"`
	Helpers_identity                                   []Foo_sub7 `json:"./helpers/identity"`
	Helpers_inherits                                   []Foo_sub7 `json:"./helpers/inherits"`
	Helpers_inheritsLoose                              []Foo_sub7 `json:"./helpers/inheritsLoose"`
	Helpers_initializerDefineProperty                  []Foo_sub7 `json:"./helpers/initializerDefineProperty"`
	Helpers_initializerWarningHelper                   []Foo_sub7 `json:"./helpers/initializerWarningHelper"`
	Helpers_instanceof                                 []Foo_sub7 `json:"./helpers/instanceof"`
	Helpers_interopRequireDefault                      []Foo_sub7 `json:"./helpers/interopRequireDefault"`
	Helpers_interopRequireWildcard                     []Foo_sub7 `json:"./helpers/interopRequireWildcard"`
	Helpers_isNativeFunction                           []Foo_sub7 `json:"./helpers/isNativeFunction"`
	Helpers_isNativeReflectConstruct                   []Foo_sub7 `json:"./helpers/isNativeReflectConstruct"`
	Helpers_iterableToArray                            []Foo_sub7 `json:"./helpers/iterableToArray"`
	Helpers_iterableToArrayLimit                       []Foo_sub7 `json:"./helpers/iterableToArrayLimit"`
	Helpers_iterableToArrayLimitLoose                  []Foo_sub7 `json:"./helpers/iterableToArrayLimitLoose"`
	Helpers_jsx                                        []Foo_sub7 `json:"./helpers/jsx"`
	Helpers_maybeArrayLike                             []Foo_sub7 `json:"./helpers/maybeArrayLike"`
	Helpers_newArrowCheck                              []Foo_sub7 `json:"./helpers/newArrowCheck"`
	Helpers_nonIterableRest                            []Foo_sub7 `json:"./helpers/nonIterableRest"`
	Helpers_nonIterableSpread                          []Foo_sub7 `json:"./helpers/nonIterableSpread"`
	Helpers_objectDestructuringEmpty                   []Foo_sub7 `json:"./helpers/objectDestructuringEmpty"`
	Helpers_objectSpread                               []Foo_sub7 `json:"./helpers/objectSpread"`
	Helpers_objectSpread2                              []Foo_sub7 `json:"./helpers/objectSpread2"`
	Helpers_objectWithoutProperties                    []Foo_sub7 `json:"./helpers/objectWithoutProperties"`
	Helpers_objectWithoutPropertiesLoose               []Foo_sub7 `json:"./helpers/objectWithoutPropertiesLoose"`
	Helpers_possibleConstructorReturn                  []Foo_sub7 `json:"./helpers/possibleConstructorReturn"`
	Helpers_readOnlyError                              []Foo_sub7 `json:"./helpers/readOnlyError"`
	Helpers_regeneratorRuntime                         []Foo_sub7 `json:"./helpers/regeneratorRuntime"`
	Helpers_set                                        []Foo_sub7 `json:"./helpers/set"`
	Helpers_setPrototypeOf                             []Foo_sub7 `json:"./helpers/setPrototypeOf"`
	Helpers_skipFirstGeneratorNext                     []Foo_sub7 `json:"./helpers/skipFirstGeneratorNext"`
	Helpers_slicedToArray                              []Foo_sub7 `json:"./helpers/slicedToArray"`
	Helpers_slicedToArrayLoose                         []Foo_sub7 `json:"./helpers/slicedToArrayLoose"`
	Helpers_superPropBase                              []Foo_sub7 `json:"./helpers/superPropBase"`
	Helpers_taggedTemplateLiteral                      []Foo_sub7 `json:"./helpers/taggedTemplateLiteral"`
	Helpers_taggedTemplateLiteralLoose                 []Foo_sub7 `json:"./helpers/taggedTemplateLiteralLoose"`
	Helpers_tdz                                        []Foo_sub7 `json:"./helpers/tdz"`
	Helpers_temporalRef                                []Foo_sub7 `json:"./helpers/temporalRef"`
	Helpers_temporalUndefined                          []Foo_sub7 `json:"./helpers/temporalUndefined"`
	Helpers_toArray                                    []Foo_sub7 `json:"./helpers/toArray"`
	Helpers_toConsumableArray                          []Foo_sub7 `json:"./helpers/toConsumableArray"`
	Helpers_toPrimitive                                []Foo_sub7 `json:"./helpers/toPrimitive"`
	Helpers_toPropertyKey                              []Foo_sub7 `json:"./helpers/toPropertyKey"`
	Helpers_typeof                                     []Foo_sub7 `json:"./helpers/typeof"`
	Helpers_unsupportedIterableToArray                 []Foo_sub7 `json:"./helpers/unsupportedIterableToArray"`
	Helpers_using                                      []Foo_sub7 `json:"./helpers/using"`
	Helpers_wrapAsyncGenerator                         []Foo_sub7 `json:"./helpers/wrapAsyncGenerator"`
	Helpers_wrapNativeSuper                            []Foo_sub7 `json:"./helpers/wrapNativeSuper"`
	Helpers_wrapRegExp                                 []Foo_sub7 `json:"./helpers/wrapRegExp"`
	Helpers_writeOnlyError                             []Foo_sub7 `json:"./helpers/writeOnlyError"`
	Package                                            string     `json:"./package"`
	Package_json                                       string     `json:"./package.json"`
	Regenerator                                        string     `json:"./regenerator"`
	Regenerator                                        string     `json:"./regenerator/"`
	Regenerator___js                                   string     `json:"./regenerator/*.js"`
}

type Foo_sub39 struct {
	Hooks Foo_sub38 `json:"hooks"`
}

type Foo_sub72 struct {
	Hooks Foo_sub71 `json:"hooks"`
}

type Foo_sub197 struct {
	Ignore  []string   `json:"ignore"`
	Linters Foo_sub196 `json:"linters"`
}

type Foo_sub65 struct {
	Ignore []string `json:"ignore"`
}

type Foo_sub196 struct {
	Js []string `json:"*.@(js)"`
	Ts []string `json:"*.@(ts)"`
}

type Foo_sub43 struct {
	Js []string `json:"*.js"`
	Md []string `json:"*.md"`
}

type Foo_sub241 struct {
	Jshint string `json:"jshint"`
	Mocha  string `json:"mocha"`
}

type Foo_sub151 struct {
	Lib_extend_node bool `json:"./lib/extend-node"`
	Lib_streams     bool `json:"./lib/streams"`
}

type Foo_sub242 struct {
	Lint string `json:"lint"`
	Test string `json:"test"`
}

type Foo_sub173 struct {
	Locate_path Foo_sub171 `json:"locate-path"`
	Path_exists Foo_sub172 `json:"path-exists"`
}

type Foo_sub169 struct {
	Locate_path string `json:"locate-path"`
	Path_exists string `json:"path-exists"`
}

type Foo_sub107 struct {
	Mail string `json:"mail"`
	URL  string `json:"url"`
}

type Foo_sub167 struct {
	Make string `json:"make"`
	Test string `json:"test"`
}

type Foo_sub283 struct {
	Microsoft_typescript_etw bool `json:"@microsoft/typescript-etw"`
	Buffer                   bool `json:"buffer"`
	Crypto                   bool `json:"crypto"`
	Fs                       bool `json:"fs"`
	Inspector                bool `json:"inspector"`
	Os                       bool `json:"os"`
	Path                     bool `json:"path"`
	Source_map_support       bool `json:"source-map-support"`
}

type Foo_sub198 struct {
	Migrations string `json:"migrations"`
}

type Foo_sub227 struct {
	Mimic_fn Foo_sub172 `json:"mimic-fn"`
}

type Foo_sub226 struct {
	Mimic_fn string `json:"mimic-fn"`
}

type Foo_sub277 struct {
	Mocha   string `json:"mocha"`
	Request string `json:"request"`
}

type Foo_sub191 struct {
	Mocha string `json:"mocha"`
}

type Foo_sub62 struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Foo_sub148 struct {
	Ndt Foo_sub147 `json:"ndt"`
}

type Foo_sub287 struct {
	Node string `json:"node"`
	Npm  string `json:"npm"`
}

type Foo_sub6 struct {
	Node string `json:"node"`
}

type Foo_sub195 struct {
	Npm string `json:"npm"`
}

type Foo_sub251 struct {
	Onetime     Foo_sub247 `json:"onetime"`
	Signal_exit Foo_sub250 `json:"signal-exit"`
}

type Foo_sub112 struct {
	Onetime     string `json:"onetime"`
	Signal_exit string `json:"signal-exit"`
}

type Foo_sub57 struct {
	Optimization_test string `json:"optimization-test"`
	Test              string `json:"test"`
}

type Foo_sub51 struct {
	Optional bool `json:"optional"`
}

type Foo_sub275 struct {
	Os_tmpdir Foo_sub274 `json:"os-tmpdir"`
}

type Foo_sub155 struct {
	Os_tmpdir string `json:"os-tmpdir"`
}

type Foo_sub234 struct {
	P_limit Foo_sub233 `json:"p-limit"`
}

type Foo_sub213 struct {
	P_limit string `json:"p-limit"`
}

type Foo_sub216 struct {
	P_locate Foo_sub215 `json:"p-locate"`
}

type Foo_sub170 struct {
	P_locate string `json:"p-locate"`
}

type Foo_sub230 struct {
	P_try Foo_sub172 `json:"p-try"`
}

type Foo_sub229 struct {
	P_try string `json:"p-try"`
}

type Foo_sub15 struct {
	Path string `json:"path"`
}

type Foo_sub208 struct {
	Pegjs string `json:"pegjs"`
}

type Foo_sub32 struct {
	Postpublish    string `json:"postpublish"`
	Posttest       string `json:"posttest"`
	PrepublishOnly string `json:"prepublishOnly"`
	Test           string `json:"test"`
}

type Foo_sub249 struct {
	Postversion    string `json:"postversion"`
	PrepublishOnly string `json:"prepublishOnly"`
	Preversion     string `json:"preversion"`
	Snap           string `json:"snap"`
	Test           string `json:"test"`
}

type Foo_sub71 struct {
	Pre_commit string `json:"pre-commit"`
}

type Foo_sub177 struct {
	Prepare    string `json:"prepare"`
	Test       string `json:"test"`
	Test_debug string `json:"test:debug"`
}

type Foo_sub278 struct {
	Prepublish string `json:"prepublish"`
	Pretest    string `json:"pretest"`
	Test       string `json:"test"`
}

type Foo_sub245 struct {
	Pretest string `json:"pretest"`
	Release string `json:"release"`
	Test    string `json:"test"`
}

type Foo_sub91 struct {
	Pretest string `json:"pretest"`
	Test    string `json:"test"`
}

type Foo_sub59 struct {
	Prettier string `json:"prettier"`
}

type Foo_sub5 struct {
	Regenerator_runtime Foo_sub4 `json:"regenerator-runtime"`
}

type Foo_sub2 struct {
	Regenerator_runtime string `json:"regenerator-runtime"`
}

type Foo_sub109 struct {
	Release string `json:"release"`
	Test    string `json:"test"`
}

type Foo_sub115 struct {
	Restore_cursor Foo_sub114 `json:"restore-cursor"`
}

type Foo_sub111 struct {
	Restore_cursor string `json:"restore-cursor"`
}

type Foo_sub26 struct {
	Rules Foo_sub25 `json:"rules"`
}

type Foo_sub80 struct {
	Rules Foo_sub79 `json:"rules"`
}

type Foo_sub93 struct {
	Rules Foo_sub92 `json:"rules"`
}

type Foo_sub182 struct {
	Safer_buffer Foo_sub181 `json:"safer-buffer"`
}

type Foo_sub150 struct {
	Safer_buffer string `json:"safer-buffer"`
}

type Foo_sub97 struct {
	Screenshot string `json:"screenshot"`
	Test       string `json:"test"`
}

type Foo_sub76 struct {
	Sindresorhus_tsconfig string `json:"@sindresorhus/tsconfig"`
	Expect_type           string `json:"expect-type"`
	Tsd                   string `json:"tsd"`
	Typescript            string `json:"typescript"`
	Xo                    string `json:"xo"`
}

type Foo_sub179 struct {
	Standard string `json:"standard"`
	Tape     string `json:"tape"`
}

type Foo_sub130 struct {
	String_width Foo_sub124 `json:"string-width"`
	Strip_ansi   Foo_sub126 `json:"strip-ansi"`
	Wrap_ansi    Foo_sub129 `json:"wrap-ansi"`
}

type Foo_sub120 struct {
	String_width string `json:"string-width"`
	Strip_ansi   string `json:"strip-ansi"`
	Wrap_ansi    string `json:"wrap-ansi"`
}

type Foo_sub189 struct {
	Tap string `json:"tap"`
}

type Foo_sub86 struct {
	Test           string `json:"test"`
	View_supported string `json:"view-supported"`
}

type Foo_sub46 struct {
	Test string `json:"test"`
}

type Foo_sub292 struct {
	Tr46               Foo_sub290 `json:"tr46"`
	Webidl_conversions Foo_sub291 `json:"webidl-conversions"`
}

type Foo_sub220 struct {
	Tr46               string `json:"tr46"`
	Webidl_conversions string `json:"webidl-conversions"`
}

type Foo_sub282 struct {
	Tsc      string `json:"tsc"`
	Tsserver string `json:"tsserver"`
}

type Foo_sub113 struct {
	Tsd string `json:"tsd"`
	Xo  string `json:"xo"`
}

type Foo_sub257 struct {
	Tslib Foo_sub256 `json:"tslib"`
}

type Foo_sub193 struct {
	Tslib string `json:"tslib"`
}

type Foo_sub3 struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Foo_sub82 struct {
	Type_fest Foo_sub81 `json:"type-fest"`
}

type Foo_sub75 struct {
	Type_fest string `json:"type-fest"`
}

type Foo_sub160 struct {
	Types_chai    string `json:"@types/chai"`
	Types_chardet string `json:"@types/chardet"`
	Types_mocha   string `json:"@types/mocha"`
	Types_node    string `json:"@types/node"`
	Types_tmp     string `json:"@types/tmp"`
	Chai          string `json:"chai"`
	Es6_shim      string `json:"es6-shim"`
	Mocha         string `json:"mocha"`
	Ts_node       string `json:"ts-node"`
	Tslint        string `json:"tslint"`
	Typescript    string `json:"typescript"`
}

type Foo_sub63 struct {
	Types_chai                      string `json:"@types/chai"`
	Types_decamelize                string `json:"@types/decamelize"`
	Types_mocha                     string `json:"@types/mocha"`
	Types_node                      string `json:"@types/node"`
	Typescript_eslint_eslint_plugin string `json:"@typescript-eslint/eslint-plugin"`
	Typescript_eslint_parser        string `json:"@typescript-eslint/parser"`
	C8                              string `json:"c8"`
	Chai                            string `json:"chai"`
	Chalk                           string `json:"chalk"`
	Coveralls                       string `json:"coveralls"`
	Cpr                             string `json:"cpr"`
	Cross_spawn                     string `json:"cross-spawn"`
	Es6_promise                     string `json:"es6-promise"`
	Eslint                          string `json:"eslint"`
	Eslint_plugin_import            string `json:"eslint-plugin-import"`
	Eslint_plugin_node              string `json:"eslint-plugin-node"`
	Gts                             string `json:"gts"`
	Hashish                         string `json:"hashish"`
	Mocha                           string `json:"mocha"`
	Rimraf                          string `json:"rimraf"`
	Standardx                       string `json:"standardx"`
	Typescript                      string `json:"typescript"`
	Which                           string `json:"which"`
	Yargs_test_extends              string `json:"yargs-test-extends"`
}

type Foo_sub176 struct {
	Types_chai              string `json:"@types/chai"`
	Types_ensure_posix_path string `json:"@types/ensure-posix-path"`
	Types_mocha             string `json:"@types/mocha"`
	Types_node              string `json:"@types/node"`
	Chai                    string `json:"chai"`
	Ensure_posix_path       string `json:"ensure-posix-path"`
	Mocha                   string `json:"mocha"`
	Typescript              string `json:"typescript"`
}

type Foo_sub96 struct {
	Types_color_convert string `json:"@types/color-convert"`
	Ava                 string `json:"ava"`
	Svg_term_cli        string `json:"svg-term-cli"`
	Tsd                 string `json:"tsd"`
	Xo                  string `json:"xo"`
}

type Foo_sub83 struct {
	Types_node string `json:"@types/node"`
	Ava        string `json:"ava"`
	Tsd        string `json:"tsd"`
	Xo         string `json:"xo"`
}

type Foo_sub285 struct {
	Typescript string `json:"typescript@*"`
}

type Foo_sub79 struct {
	Typescript_eslint_ban_types              string `json:"@typescript-eslint/ban-types"`
	Typescript_eslint_indent                 string `json:"@typescript-eslint/indent"`
	Node_no_unsupported_features_es_builtins string `json:"node/no-unsupported-features/es-builtins"`
}

type Foo_sub25 struct {
	Typescript_eslint_member_ordering string `json:"@typescript-eslint/member-ordering"`
	No_redeclare                      string `json:"no-redeclare"`
	Unicorn_better_regex              string `json:"unicorn/better-regex"`
	Unicorn_prefer_includes           string `json:"unicorn/prefer-includes"`
	Unicorn_prefer_string_slice       string `json:"unicorn/prefer-string-slice"`
	Unicorn_string_content            string `json:"unicorn/string-content"`
}

type Foo_sub14 struct {
	URL string `json:"url"`
}

type Foo_sub49 struct {
	Ungap_url_search_params                          string `json:"@ungap/url-search-params"`
	Abort_controller                                 string `json:"abort-controller"`
	Abortcontroller_polyfill                         string `json:"abortcontroller-polyfill"`
	Babel_core                                       string `json:"babel-core"`
	Babel_plugin_istanbul                            string `json:"babel-plugin-istanbul"`
	Babel_plugin_transform_async_generator_functions string `json:"babel-plugin-transform-async-generator-functions"`
	Babel_polyfill                                   string `json:"babel-polyfill"`
	Babel_preset_env                                 string `json:"babel-preset-env"`
	Babel_register                                   string `json:"babel-register"`
	Chai                                             string `json:"chai"`
	Chai_as_promised                                 string `json:"chai-as-promised"`
	Chai_iterator                                    string `json:"chai-iterator"`
	Chai_string                                      string `json:"chai-string"`
	Codecov                                          string `json:"codecov"`
	Cross_env                                        string `json:"cross-env"`
	Form_data                                        string `json:"form-data"`
	Is_builtin_module                                string `json:"is-builtin-module"`
	Mocha                                            string `json:"mocha"`
	Nyc                                              string `json:"nyc"`
	Parted                                           string `json:"parted"`
	Promise                                          string `json:"promise"`
	Resumer                                          string `json:"resumer"`
	Rollup                                           string `json:"rollup"`
	Rollup_plugin_babel                              string `json:"rollup-plugin-babel"`
	String_to_arraybuffer                            string `json:"string-to-arraybuffer"`
	Teeny_request                                    string `json:"teeny-request"`
}

type Foo_sub147 struct {
	Versions []string `json:"versions"`
}

type Foo_sub156 struct {
	Vows string `json:"vows"`
}

type Foo_sub48 struct {
	Whatwg_url string `json:"whatwg-url"`
}

type Foo_sub255 struct {
	_ Foo_sub254 `json:"."`
	_ string     `json:"./"`
}

type Foo_sub77 struct {
	_ []string `json:"*"`
}

type Foo_sub1 struct{}
