---
id: TypeScript_Cheat_Sheet
aliases:
  - TypeScript
  - TS Cheat Sheet
tags:
  - coding
  - typescript
  - cheatsheet
dg-publish: true
---

# TypeScript Quick Cheat Sheet

> Quick reference guide for TypeScript types, syntax, generics, and utility patterns.

### 🔗 Related Vault Notes
- [[Basics|JS Basics]]: Core JavaScript syntax and scoping
- [[Arrow Functions|JS Arrow Functions]]: Arrow function syntax and lexical `this`
- [[Classes|JS Classes]]: OOP classes and prototypes
- [[Objects|JS Objects]]: Object properties and methods
- [[Asyncronous Programming|JS Async]]: Promises, async/await, and event loop
- [[React Cheat Sheet]]: React + TypeScript patterns, typed props, hooks, and events
- [[SOLID principle]]: Object-oriented design principles in TypeScript

---

## 1. Primitive & Basic Types

```ts
let isDone: boolean = false;
let count: number = 42;
let username: string = "Arun";
let notSure: any = 4;              // Disables type checking (avoid when possible)
let unknownVal: unknown = 4;       // Safe alternative to 'any' - requires narrowing before use
let unusable: void = undefined;     // Typically for functions returning nothing
let u: undefined = undefined;
let n: null = null;
let neverVal: never;               // Values that never occur (e.g. infinite loops, throw)
```

> [!TIP]
> Prefer `unknown` over `any`. With `unknown`, TypeScript forces you to check the type using type guards before performing operations on the value.

---

## 2. Arrays, Tuples & Enums

```ts
// Arrays
let list: number[] = [1, 2, 3];
let genericList: Array<string> = ["a", "b", "c"];
let readonlyList: readonly number[] = [1, 2, 3];

// Tuples: Fixed-length and specific types per position
let tuple: [string, number] = ["status", 200];
let [statusMsg, statusCode] = tuple;

// Enums
enum Direction {
  Up = 1,
  Down,
  Left,
  Right,
}
let dir: Direction = Direction.Up;

// Const Enum (inlines values at compile time for zero runtime overhead)
const enum HttpStatus {
  OK = 200,
  NotFound = 404,
  InternalError = 500,
}
```

---

## 3. Type Aliases vs Interfaces

| Feature | `interface` | `type` |
| :--- | :--- | :--- |
| **Objects & Methods** | Yes | Yes |
| **Primitives, Unions, Tuples** | No | Yes |
| **Extending** | `extends` | `&` (Intersection) |
| **Declaration Merging** | Yes (merges multiple declarations) | No (duplicate identifier error) |

```ts
// Interface: Great for object shapes, API responses, class contracts
interface User {
  readonly id: string;             // Readonly property
  name: string;
  age?: number;                    // Optional property
  sayHello(): string;              // Method signature
}

// Interface Extension
interface AdminUser extends User {
  role: "admin" | "superadmin";
  permissions: string[];
}

// Type Alias: Great for unions, primitives, tuples, utility types
type Point = {
  x: number;
  y: number;
};

type ID = string | number;         // Union type
type Status = "idle" | "loading" | "success" | "error"; // String literal union
```

---

## 4. Union & Intersection Types

```ts
// Union (|): Value can be one of several types
type Result = SuccessResponse | ErrorResponse;

function printId(id: string | number) {
  if (typeof id === "string") {
    console.log(id.toUpperCase());
  } else {
    console.log(id.toFixed(2));
  }
}

// Intersection (&): Combines multiple types into one
type Person = { name: string };
type Employee = { employeeId: number };
type Staff = Person & Employee; // Must have both 'name' and 'employeeId'
```

---

## 5. Functions & Signatures

> Related: [[Arrow Functions]], [[Immediately invoked functions]]

```ts
// Named Function with parameter & return type
function add(a: number, b: number): number {
  return a + b;
}

// Arrow Function type definition
type MathOp = (x: number, y: number) => number;
const multiply: MathOp = (x, y) => x * y;

// Optional & Default Parameters
function greet(name: string, greeting: string = "Hello", title?: string): string {
  return `${greeting} ${title ? title + " " : ""}${name}!`;
}

// Rest Parameters
function sum(...numbers: number[]): number {
  return numbers.reduce((acc, curr) => acc + curr, 0);
}

// Function Overloads
function parseInput(input: string): string[];
function parseInput(input: number): number[];
function parseInput(input: string | number): any {
  if (typeof input === "string") return input.split("");
  return [input];
}
```

---

## 6. Generics

```ts
// Generic Function
function identity<T>(arg: T): T {
  return arg;
}
const num = identity<number>(42);
const str = identity("auto-inferred");

// Generic Constraints (extends)
interface HasLength {
  length: number;
}
function logLength<T extends HasLength>(item: T): number {
  return item.length; // Safe because T is guaranteed to have .length
}

// Generic Interface
interface ApiResponse<TData> {
  data: TData;
  status: number;
  timestamp: string;
}

type UserResponse = ApiResponse<User>;
```

---

## 7. Type Narrowing & Guards

```ts
// 1. typeof
function processValue(x: string | number) {
  if (typeof x === "string") {
    return x.trim();
  }
  return x * 10;
}

// 2. instanceof (see [[Classes]])
function handleDate(d: Date | string) {
  if (d instanceof Date) {
    return d.toISOString();
  }
  return d;
}

// 3. 'in' operator
type Cat = { meow: () => void };
type Dog = { bark: () => void };

function speak(animal: Cat | Dog) {
  if ("meow" in animal) {
    animal.meow();
  } else {
    animal.bark();
  }
}

// 4. Custom Type Guard (is operator)
function isCat(animal: Cat | Dog): animal is Cat {
  return (animal as Cat).meow !== undefined;
}

// 5. Discriminated Union (Best practice for state machines)
type NetworkState =
  | { state: "loading" }
  | { state: "failed"; code: number; message: string }
  | { state: "success"; response: { title: string } };

function renderState(s: NetworkState) {
  switch (s.state) {
    case "loading":
      return "Loading...";
    case "failed":
      return `Error ${s.code}: ${s.message}`;
    case "success":
      return `Title: ${s.response.title}`;
  }
}
```

---

## 8. Essential Utility Types

TypeScript comes with built-in type transformers:

```ts
interface Todo {
  id: string;
  title: string;
  completed: boolean;
  dueDate: Date;
}

// Partial<T>: Makes all properties optional
type UpdateTodoInput = Partial<Todo>;

// Required<T>: Makes all properties required
type StrictTodo = Required<Todo>;

// Readonly<T>: Makes all properties immutable
type ReadonlyTodo = Readonly<Todo>;

// Pick<T, K>: Constructs type by picking specified keys
type TodoPreview = Pick<Todo, "id" | "title">;

// Omit<T, K>: Constructs type by removing specified keys
type CreateTodoInput = Omit<Todo, "id">;

// Record<K, T>: Constructs object type with keys K and value type T
type TodoStatusMap = Record<string, Todo>;

// ReturnType<T>: Obtains return type of a function type
function createSession() {
  return { token: "xyz", expiresAt: Date.now() };
}
type Session = ReturnType<typeof createSession>;

// Awaited<T>: Unwraps Promise types (see [[Asyncronous Programming]])
type AsyncData = Awaited<Promise<string>>; // string
```

---

## 9. Key Operators & Type Assertions

```ts
// keyof: Produces union of keys of an object type
type TodoKeys = keyof Todo; // "id" | "title" | "completed" | "dueDate"

// Indexed Access Type
type TitleType = Todo["title"]; // string

// as const: Creates deeply readonly literal types
const CONFIG = {
  endpoint: "https://api.example.com",
  timeout: 5000,
} as const;
// CONFIG.endpoint is literal "https://api.example.com" (not string), readonly

// Non-null Assertion Operator (!) - Assert value is not null/undefined
const elem = document.getElementById("root")!;

// Type Assertion (as)
const rawData: unknown = '{"id": 1}';
const parsed = rawData as string;
```

---

## 10. Classes & Modifiers

> Related: [[Classes]], [[SOLID principle]], [[Repository Design Pattern]]

```ts
class BaseService {
  public id: string;               // Accessible anywhere (default)
  protected createdAt: Date;       // Accessible in class and subclasses
  private secretKey: string;       // Accessible only inside BaseService
  readonly version = "1.0.0";      // Cannot be modified after instantiation

  constructor(id: string, secret: string) {
    this.id = id;
    this.secretKey = secret;
    this.createdAt = new Date();
  }
}

// Parameter Properties (Constructor shorthand)
class DatabaseService {
  constructor(
    public readonly host: string,
    private port: number,
    private sslEnabled: boolean = true
  ) {}
}
```

---

## 11. React + TypeScript Quick Reference

> See complete guide in [[React Cheat Sheet]]

```tsx
// Typing Component Props
interface ButtonProps {
  label: string;
  onClick: (event: React.MouseEvent<HTMLButtonElement>) => void;
  variant?: "primary" | "secondary";
  children?: React.ReactNode;
}

export const Button = ({ label, onClick, variant = "primary", children }: ButtonProps) => {
  return (
    <button className={`btn-${variant}`} onClick={onClick}>
      {label}
      {children}
    </button>
  );
};
```
