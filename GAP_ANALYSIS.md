# Gap Analysis: Implementation vs Architecture Spec

## M1: Runtime Cleanup
| Spec Requires | Current State | Gap |
|---|---|---|
| `Runtime` interface: `Start/Stop/Execute/Events/Capabilities/Agents` | `Runtime` struct with methods ✅ | Нужно выделить interface |

## M2: Agent Runtime
| Spec Requires | Current State | Gap |
|---|---|---|
| `Agent` interface: `ID() string`, `Metadata() Metadata`, `Execute(AgentContext) error` | `Agent` struct с `ID()`, `Info()`, state machine ✅ | Нет `Execute(AgentContext)` метода |

## M3: Capability System
| Spec Requires | Current State | Gap |
|---|---|---|
| `CapabilityRequest{Name, Input, Context}` | Свои типы (`Request` в runtime) ⚠️ | Нужно выровнять с SDK |
| `CapabilityResult{Output, Error}` | Есть адаптер ✅ | OK |

## M4: Event Bus
| Spec Requires | Current State | Gap |
|---|---|---|
| `EventBus` interface: `Publish(ctx, Event)`, `Subscribe(Subscription)`, `Unsubscribe(string)` | `Bus` struct, другой API ⚠️ | Нужно выровнять интерфейс |

## M5: Policy Engine
| Spec Requires | Current State | Gap |
|---|---|---|
| `PolicyEngine.Evaluate(ctx, CapabilityRequest)` -> `Decision` | `Engine.Evaluate(ctx, Request)` -> `Result` ⚠️ | Нужно выровнять типы |

## M6: Transport
| Spec Requires | Current State | Gap |
|---|---|---|
| `Transport` interface: `Name()/Start/Stop/Send/Peers()` | Транспорт есть, сигнатуры отличаются ⚠️ | Нужно выровнять |

## M7-M15: Not Started
| Milestone | Status |
|---|---|
| M7: Reticulum Transport | ❌ |
| M8: Distributed Runtime | ❌ |
| M9: Memory (4-layer) | Базовый in-memory есть, spec требует Working/Session/Persistent/External |
| M10: Scheduler | ❌ |
| M11: Plugin SDK | Базовый `sdk/plugin/` есть, spec требует больше |
| M12: Planner Interface | ✅ Базовый `sdk/planner/` есть, но нужно выровнять |
| M13: Mesh Services | ❌ |
| M14: Security Hardening | ❌ |
| M15: API Freeze | ❌ |
