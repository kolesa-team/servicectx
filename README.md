## xoptions: custom context propagation across microservices through HTTP headers and/or OpenTelemetry

[![Actions Status](https://github.com/kolesa-team/xoptions/workflows/test/badge.svg)](https://github.com/kolesa-team/xoptions/actions)
[![codecov](https://codecov.io/gh/kolesa-team/xoptions/branch/main/graph/badge.svg?token=j7K2w57hif)](https://codecov.io/gh/kolesa-team/xoptions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kolesa-team/xoptions)](https://goreportcard.com/report/github.com/kolesa-team/xoptions)

A common issue in (micro)services architecture is exchanging and overriding some arbitrary options across the service chain.
For example, if the project calls a billing service at `http://billing-default`, then how can we test its new version at `http://billing-v2`? 
One way is to manually update project configuration, then build and deploy it, as well as every other project that calls billing service. 
That is inconvenient, especially if the call chain contains multiple links, or if there are many projects that depend on the billing service.

One solution would be to pass the options across services through HTTP headers like `x-service-{SERVICE_NAME}-{OPTION}: {VALUE}`. 
For example, if we want service `A` to ask service `B` to use a custom branch of service `C`, we can do that with a header `x-service-c-branch: my-branch`. 
When processing a request, service `B` must parse this header, apply its value within itself (say, reconfigure an HTTP client), and then also pass that header to every other service.
This is not limited to branch names or URLs; any arbitrary option can be passed through such a header.

The library aims to standardize and automate this workflow. 
It handles parsing, reading and writing options to and from

* custom HTTP headers (e.g. `x-service-api-branch: feature-123`)
* query strings (e.g. `?x-service-api-version=2`)
* OpenTelemetry or OpenTracing baggages
* ...and `context.Context`

This library is inspired in part by the article from DoorDash on [OpenTelemetry for custom context propagation](https://doordash.engineering/2021/06/17/leveraging-opentelemetry-for-custom-context-propagation/).

### Usage

#### Retrieving options from request headers
```go
import "github.com/kolesa-team/xoptions"

func testHandler(w http.ResponseWriter, r *http.Request) {
	options := xoptions.FromHeaders(r.Header)

	// read an API version from header, or use 1.0 by default
	apiVersion := options.Get("api", "version", "1.0")
	w.Write([]byte(fmt.Sprintf("API version: %s\n")))
}

// Output with no extra headers:
// API version: 1.0

// Output with "x-service-api-version: 2.1" header
// API version: 2.1
```

#### Passing options through context

```go
import "github.com/kolesa-team/xoptions"

func testHandlerWithContext(w http.ResponseWriter, r *http.Request) {
	// parse options from headers and add them to a context.
	// it's ok if no special headers were sent: an empty struct is then used instead.
	ctx := xoptions.InjectIntoContextFromHeaders(r.Context(), r.Header)

	// a remoteCall is probably defined in another package;
	// its `username` argument is a part of business logic,
	// but inter-service options are passed in `ctx` as an ancillary data.
	remoteCall := func(ctx context.Context, username string) string {
            // options are retrieved from a context
            opts := FromContext(ctx)
            // the remote API address is taken from these options (or default URL is used instead).
            url := opts.Get("api", "url", "http://api")
            url += "?username=" + username
            apiRequest, _ := http.NewRequest("GET", url, nil)
            // the options are propagated further within the headers
            opts.InjectIntoHeaders(apiRequest.Header)
            // TODO: execute remote call
            // _, _ = http.DefaultClient.Do(apiRequest)
            
            return fmt.Sprintf("Calling remote API at %s with headers:\n%+v", url, apiRequest.Header)
	}

	w.Write([]byte(remoteCall(ctx, r.URL.Query().Get("username"))))
}
```

```shell
$ curl --header "x-service-api-url: http://my-custom-api" --header "x-service-billing-branch: hotfix-123" http://localhost?username=Mary
Calling remote API at http://my-custom-api?username=Alex with headers:
map[X-Service-Api-Url:[http://my-custom-api] X-Service-Billing-Branch:[hotfix-123]]
```

#### Replacing branch name in URL

One typical use-case is dynamically replacing a branch name in a URL. The library offers a helper function to make URLs easily configurable:

* Say, the project calls `http://billing-main`, where `main` is a branch name. Then use the address `http://billing-$branch` internally instead. Here, `$branch` is a placeholder to be replaced.
* Use `x-service-billing-branch: my-branch` header and call `ReplaceUrlBranch` helper function to configure a URL:
```go
import "github.com/kolesa-team/xoptions"

func testHandler(w http.ResponseWriter, r *http.Request) {
	options := xoptions.FromHeaders(r.Header)

	// retrieve a `billing` service branch, or use `main` by  default
	billingBranch := options.Get("billing", "branch", "main")
	// replace `$branch` with billingBranch
	billingUrl := xoptions.ReplaceUrlBranch("http://billing-$branch", billingBranch)
	
	fmt.Println(billingUrl)	
	// curl --header "x-service-billing-branch: bugfix-123" http://localhost
	// -> http://billing-bugfix-123
	
}
```

#### Interacting with options manually

```
// create a set of options
options := xoptions.New()
options.Set("api", "branch", "feature-123")

// retrieve them as a map of HTTP headers
fmt.Printf("%+v", options.HeaderMap())
// map[x-service-api-url:feature-123]

// read integer value (or use 1 as a default)
options.Set("api", "version", "2")
fmt.Println(options.GetInt("api", "version", 1))
// 2

// read time.Duration (or use 1 second as a default)
options.Set("api", "timeout", "3s")
fmt.Println(options.GetDuration("api", "timeout", time.Second))
// 3s
```

### Advantages

* A simple option format. `x-service-{SERVICE_NAME}-{OPTION}` is trivially parsed in any programming language (if your services are written in other languages).
* Can be used with [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) or [OpenTracing](https://github.com/opentracing/opentracing-go).
* No external dependencies.

### Limitations

* Service names in `x-service-{SERVICE_NAME}-{OPTION}` must not contain `-` sign (which is used as a separator).
The format is not configurable for the sake of simplicity.
* The library can't "un-hardcode" your project configuration automagically. Overriding some options (such as URLs) per request is trivial in application code, and some (like database hosts) is not.
* Clearly, accepting configuration from arbitrary headers is a security violation. An application code is responsible for disabling this functionality in production.

## xoptions: библиотека для обмена опциями при межсервисном взаимодействии через заголовки HTTP

При разработке и тестировании в микросервисной архитектуре часто возникает задача передачи и переопределения опций по цепочке сервисов. Например, если проект обращается к сервису биллинга по адресу `http://billing-default`, то как можно протестировать новую версию по адресу `http://billing-v2`? Чаще всего приходится изменять конфигурацию зависимого проекта вручную, запускать ради этого сборку и деплой. Это неудобно, особенно если цепочка вызовов состоит более чем из 2 проектов. 

Вариант решения: передавать опции между сервисами через заголовки в формате `x-service-{SERVICE_NAME}-{OPTION}: {VALUE}`. Например, чтобы из сервиса `A` попросить сервис `B` использовать нужную ветку сервиса `C`, можно передать заголовок `x-service-c-branch: my-branch`. Сервис `B` принимает все заголовки такого формата, применяет их внутри самого себя, а также прокидывает их далее по цепочке во все сервисы, в которые он будет обращаться.

Библиотека позволяет стандартизировать и автоматизировать парсинг заголовков такого формата, их чтение и запись в `http.Header` и передачу опций через `context`.
Фактическое применение этих опций остаётся ответственностью самих сервисов.

---

© 2022 Kolesa Group. Licensed under [MIT](https://opensource.org/licenses/MIT)
