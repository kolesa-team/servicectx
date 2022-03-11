## servicectx: custom context propagation across microservices through HTTP headers, query strings, and OpenTelemetry

[![Actions Status](https://github.com/kolesa-team/servicectx/workflows/test/badge.svg)](https://github.com/kolesa-team/servicectx/actions)
[![codecov](https://codecov.io/gh/kolesa-team/servicectx/branch/main/graph/badge.svg?token=j7K2w57hif)](https://codecov.io/gh/kolesa-team/servicectx)
[![Go Report Card](https://goreportcard.com/badge/github.com/kolesa-team/servicectx)](https://goreportcard.com/report/github.com/kolesa-team/servicectx)

A common issue in (micro)services architecture is exchanging and overriding some arbitrary properties across the service chain.
While existing tools like [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) do provide an underlying infrastructure for that, there is a lack of conventions on how to use them at the application level.

This library aims to ease the inter-service communication, development, and testing by 
* Defining a standard key format for properties meant for propagation: `x-service-{SERVICE_NAME}-{OPTION_NAME}`
  * This way it is easily distinguishable who the property is intended for.
    When a service receives such property, it can decide to act on it by reconfiguring itself (if it has something to do with `{SERVICE_NAME}`), or just pass it further.
* Providing a convenient way of passing the properties via
  * HTTP headers (e.g. `x-service-api-branch: feature-123`)
  * Query strings (e.g. `?x-service-api-version=2`)
  * OpenTelemetry/OpenTracing baggage
  * `context.Context`, within a single Go process / request handler 
  
Why would someone need that? One notable use-case is dynamic routing. 
Say you're developing a new version of a billing service `http://billing-v2` that is usually called by the backend at `http://billing-v1`. With `servicectx`, you can conveniently propagate a custom billing URL without modifying the backend or user interface:
* First, find a way to pass a custom billing URL to the backend. 
  * If using a web page, add a query parameter `?x-service-billing-url=billing-v2` or `x-service-billing-url: billing-v2` header (some browser extensions can help).
* Use `servicectx` on the backend to parse this property from query string and/or headers.
    * For example: `billingUrl := servicectx.FromRequest(req).Get("billing", "url", "billing-v1")`
    * Which means: if there's a `x-service-billing-url` header, then use its value; otherwise, use `billing-v1` by default.
* Call this billing URL instead of the hardcoded one...
* ...but also propagate all the properties received by the backend to your billing service as well. 
  * `servicectx` can inject them into OpenTelemetry/OpenTracing baggage, or just pass via HTTP headers.

In real life, service chains can be longer and more complex. What if you had to modify the configuration of 5 other services manually just to replace `http://billing-v1` with a new URL? When it comes to that, it seems reasonable to implement a common, standard solution for properties propagation across all these services. Take a look at a [more detailed example here](docs/example-usecase.jpg).

The usage is not limited to just dynamic routing. `servicectx` can help propagate any ancillary data and make any property dynamically reconfigurable (if your application code can do that, of course). How about increasing log verbosity for a specific request with, say, `x-service-api-log-level: debug`? [The possibilities in ever-growing microservices architecture seem infinite](https://www.youtube.com/watch?v=y8OnoxKotPQ).

This library is inspired in part by an article from DoorDash on [OpenTelemetry for custom context propagation](https://doordash.engineering/2021/06/17/leveraging-opentelemetry-for-custom-context-propagation/).

### Usage

#### Retrieving properties from request
```go
import "github.com/kolesa-team/servicectx"

func testHandler(w http.ResponseWriter, r *http.Request) {
	props := servicectx.FromRequest(r)

	// read an API version from request, or use 1.0 by default
	apiVersion := props.Get("api", "version", "1.0")
	fmt.Printf("API version: %s\n")
}

// Output if no extra headers were sent:
// API version: 1.0

// Output if a header "x-service-api-version: 2.1" was sent
// API version: 2.1
```

#### Passing properties via Go context

```go
import "github.com/kolesa-team/servicectx"

func testHandlerWithContext(w http.ResponseWriter, r *http.Request) {
	// parse properties from request and add them to a context.
	// it's ok if no special headers or query args were sent: an empty struct is then used instead.
	ctx := servicectx.InjectIntoContextFromRequest(r.Context(), r.Header)

	// a remoteCall is probably defined in another package;
	// its `username` argument is a part of business logic,
	// but custom context is passed in `ctx` as an ancillary data.
	remoteCall := func(ctx context.Context, username string) string {
            // options are retrieved from a context
            props := servicectx.FromContext(ctx)
            // the remote API address is taken from these props (or default URL is used instead).
            url := props.Get("api", "url", "http://api")
            url += "?username=" + username
            apiRequest, _ := http.NewRequest("GET", url, nil)
            // the properties are propagated further within the headers
            props.InjectIntoHeaders(apiRequest.Header)
            
            // TODO: execute remote call
            // _, _ = http.DefaultClient.Do(apiRequest)
            
            return fmt.Sprintf("Calling remote API at %s with headers:\n%+v", url, apiRequest.Header)
	}

	w.Write([]byte(remoteCall(ctx, r.URL.Query().Get("username"))))
}
```

Calling the handler above with curl will get us:

```shell
$ curl --header "x-service-api-url: http://my-custom-api" --header "x-service-billing-branch: hotfix-123" http://localhost?username=Mary

Calling remote API at http://my-custom-api?username=Mary with headers:
map[X-Service-Api-Url:[http://my-custom-api] X-Service-Billing-Branch:[hotfix-123]]
```

#### Dynamic routing: replacing branch name in URL

One typical use-case for custom context is a dynamic replacement of branch names in URLs. The library offers a helper function to make URLs easily configurable:

* Say, the project calls `http://billing-main` by default, where `main` is a branch name. 
* We propose to store that address as `http://billing-$branch` instead, where `$branch` is a placeholder to be replaced.
* Then use `x-service-billing-branch: my-branch` property and call `servicectx.ReplaceUrlBranch` helper function to reconfigure a URL on the fly:
```go
import "github.com/kolesa-team/servicectx"

func testHandler(w http.ResponseWriter, r *http.Request) {
    props := servicectx.FromRequest(r)
	// retrieve a `billing` service branch, or use `main` by  default
	billingBranch := props.Get("billing", "branch", "main")
	// replace `$branch` with billingBranch
	billingUrl := servicectx.ReplaceUrlBranch("http://billing-$branch", billingBranch)
	
	fmt.Println(billingUrl)	
	
	// curl --header "x-service-billing-branch: bugfix-123" http://localhost
	// -> http://billing-bugfix-123	
}
```

#### OpenTelemetry and OpenTracing

Custom properties can be written to and read from telemetry contexts.

`FromContextAndBaggage` (and its opentracing counterpart, `FromContextAndSpan`) extracts custom properties from baggage or span. `InjectIntoBaggage` (and `InjectIntoSpan`) injects them into baggage or span.

See examples in [kolesa-team/servicectx/otel](otel), [kolesa-team/servicectx/opentracing](opentracing).

#### Interacting with properties

```go
props := servicectx.New()
props.Set("api", "branch", "feature-123")

// retrieve the properties as a map of HTTP headers
fmt.Printf("%+v", props.HeaderMap())
// map[x-service-api-branch:feature-123]

// read integer value (or use 1 as a default)
props.Set("api", "version", "2")
fmt.Println(props.GetInt("api", "version", 1))
// 2

// read time.Duration (or use 1 second as a default)
props.Set("api", "timeout", "3s")
fmt.Println(props.GetDuration("api", "timeout", time.Second))
// 3s
```

### Advantages

* A simple format. `x-service-{SERVICE_NAME}-{OPTION}` can be easily parsed in any programming language, if you need it.
* Supports [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) and [OpenTracing](https://github.com/opentracing/opentracing-go)...
* ...but can also be used as a standalone solution.
* The properties from multiple sources can be merged (e.g. an HTTP header can take preference over the same property from OpenTracing baggage).
* No external dependencies (except OpenTracing/OpenTelemetry, when you need them).

### Concerns

* Service names in `x-service-{SERVICE_NAME}-{OPTION}` cannot contain `-` sign (which is used as a separator).
The format is not configurable for the sake of simplicity.
  * If someone really wants to, it is probably possible to introduce custom format without breaking the compatibility.
* The library can't "un-hardcode" your project configuration automagically. Overriding some properties per-request in application code (such as HTTP URLs) is trivial, and some (like database hosts) is not.
* Clearly, accepting arbitrary configuration from user input is a security violation. An application code is responsible for disabling this functionality in production.

### servicectx: передача контекста между сервисами через заголовки HTTP, query-параметры или OpenTelemetry

При разработке и тестировании в микросервисной архитектуре часто возникает задача передачи и переопределения произвольных опций в цепочке сервисов. Существующие решения вроде OpenTelemetry предоставляют для этого техническую инфраструктуру, но на практике ощущается недостаток *соглашений или стандартов* по их использованию в бизнес-логике.

Задачи библиотеки:

* Описать стандартный формат ключей для межсервисного взаимодействия: `x-service-{SERVICE_NAME}-{OPTION_NAME}`
    * Такой формат позволяет легко понять, для какого именно сервиса предназначен ключ. При получении ключа сервис может отреагировать на него (если у него подходящий `{SERVICE_NAME}`), переконфигурировав себя, либо просто прокинуть это свойство дальше.
* Предоставить удобные способы передачи таких данных через
    * Заголовки HTTP (например, `x-service-api-branch: feature-123`)
    * Query-параметры (например, `?x-service-api-version=2`)
    * Метаданные (baggage) OpenTelemetry/OpenTracing
    * или через `context.Context` для передачи в рамках одного процесса Go

Зачем это может понадобиться? Интересный вариант использования - это динамический роутинг. Например, мы разрабатываем новую версию сервиса платежей `http://billing-v2`, в то время как зависимый от него бэкенд обращается к `http://billing-v1`. С помощью `servicectx` можно удобно переопределить адрес биллинга, не меняя зависимые проекты и пользовательский интерфейс, чтобы быстро протестировать интеграцию всех сервисов:
* Сначала нужно передать новый адрес сервиса платежей на бэкенд
    * Если речь о веб-странице, то можно добавить в адресную строку параметр `?x-service-billing-url=billing-v2` или установить заголовок `x-service-billing-url: billing-v2` (с этим могут помочь браузерные расширения).
* Использовать `servicectx` на бэкенде для парсинга опций из запроса.
    * Например: `billingUrl := servicectx.FromRequest(req).Get("billing", "url", "billing-v1")`
    * Это означает: если пришёл заголовок `x-service-billing-url`, то используем его значение как адрес биллинга; иначе используем `billing-v1` по-умолчанию.
* Вызвать сервис платежей по этому адресу (вместо использования захардкоженного адреса).
* Также хорошим решением будет прокинуть весь контекст, полученный бэкендом из интерфейса, в сервис биллинга.
    * `servicectx` может внедрить его в трейс OpenTelemetry/OpenTracing через baggage, или просто передать в HTTP заголовках также, как это было сделано на первом шаге.
    * Это позволит контексту распространиться дальше, по всей цепочке вызовов.

В реальности цепочки вызовов сервисов бывают более длинными и сложными. Что если нам пришлось бы вручную менять конфигурацию пяти разных проектов просто чтобы заменить `http://billing-v1` на новый URL? В такой ситуации разумно внедрить общее, стандартное решение для передачи контекста между всеми проектами и избавиться от необходимости вносить изменения вручную.

Использование библиотеки не ограничивается динамическим роутингом. `servicectx` поможет принять и передать любые служебные данные и сделать любое свойство конфигурируемым (если код приложения сможет с этим работать). Например, можно реализовать изменение уровня логирования в рамках одного запроса через заголовок типа  `x-service-api-log-level: debug`.

---

© 2022 Kolesa Group. Licensed under [MIT](https://opensource.org/licenses/MIT)
