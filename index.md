
### Basic
#### Define Normal Variable
<script id="asciicast-ZpUr10sc1b9pChR2mi5YyU58x" src="https://asciinema.org/a/ZpUr10sc1b9pChR2mi5YyU58x.js" async></script>
#### Define Container Variable
<script id="asciicast-eSrgNRe0U8bXQCgYn7vWK98fC" src="https://asciinema.org/a/eSrgNRe0U8bXQCgYn7vWK98fC.js" async></script>
#### Define Reference
<script id="asciicast-TJ3WuG41dYrW0KV4CgxjR76ox" src="https://asciinema.org/a/TJ3WuG41dYrW0KV4CgxjR76ox.js" async></script>
<script id="asciicast-XDOzW4BOU5DDffkVrmV6INKFT" src="https://asciinema.org/a/XDOzW4BOU5DDffkVrmV6INKFT.js" async></script>
#### Define Function Variable
<script id="asciicast-9EOJaJn7Pp5SPJPtjZqfQYtK8" src="https://asciinema.org/a/9EOJaJn7Pp5SPJPtjZqfQYtK8.js" async></script>
#### Delete Variable / Reference
<script id="asciicast-BLCVobocSt6XKIZtfCnHyvEVq" src="https://asciinema.org/a/BLCVobocSt6XKIZtfCnHyvEVq.js" async></script>
#### Conditional Expression(Statement)
<script id="asciicast-ZI5gULz2gWRrI9OJXpY0HZrli" src="https://asciinema.org/a/ZI5gULz2gWRrI9OJXpY0HZrli.js" async></script>
#### Loop Expression(Statement)
<script id="asciicast-c9xlHx0jm0qaBh1yugIT9vESS" src="https://asciinema.org/a/c9xlHx0jm0qaBh1yugIT9vESS.js" async></script>
<script id="asciicast-NNxxZSIQvJc10ZRxhARwOTJVa" src="https://asciinema.org/a/NNxxZSIQvJc10ZRxhARwOTJVa.js" async></script>
#### Import / Export
- `let export = ...` to export variable

### Call Function
There are two ways to call a function
#### Define a function
```
let f = func(a) {
    ret a;
};
```
#### Call function via normal way
```
f(1);
```
#### Call function via simple way
```
f 1;
```
#### Types are allowed in simple way
```
f 1;
f 1.0;
f "x";
f 'c';
f void;
f true;
f false;
f {
    printLine "Hello World!";
};
```
Those types are allowed

**Important: When {} is used after a function, it means underline function, instead of hash**
```
let a = { "x": [1, 2, 3] };
f a;
```
Complex types are allowed only when they are stored at single variable
#### The limitation of simple way
```
let a = [1, 2, 3];
f a[0];
```
This does not work because the program will try to evaluate this: `f(a)[0]`
#### Simple way only pass one argument to the function
```
let a = 1;
let b = 2;
f a, b;
```
This does not work

### Native Functions (Built-in) Part 1
#### IO
- `printLine "Hello World!";` to print "Hello World!" and switch to the next line
- `print "Hello World!\n";` to print "Hello World!\n" ("\n" is new-line mark)
- `input();` to input a string end by space
- `inputLine();` to read a line
#### Type Convert
- `type 1;` to get the type of number 1 ("Integer")
- `string 12;` to convert 12 to string ("12")
- `integer "40";` to convert "50" to integer (40)
- `float 4;` to convert 4 to float (4.0)
- `boolean 3;` to convert 3 to boolean (true)
#### Array
##### append(arr, ele)
```
let a = []
a = append(a, 1)
a = append(a, 2)
printLine a
```
This will print "[1, 2]"

##### array(len, first, nextFunc)
nextFunc is defined as func(index, previousValue) { ret nextValue; }

array call nextFunc at each time to generate the next element, for the first call, previousValue = first
- `array 5;` to get \[void, void, void, void, void]
- `array(5, 0);` to get \[0, 0, 0, 0, 0]
- `array(5, 0, func(i, p) { ret p + i; });` to get \[0, 1, 3, 6, 10]
##### first / last
- `first([1, 2, 3])` to get 1
- `let a = [1, 2, 3]; first a = 4; a;` to modify the first element, a will be \[4, 2, 3]
- `last([1, 2, 3])` to get 3
- `let a = [1, 2, 3]; last a = 9; a;` to modify the first element, a will be \[1, 2, 9]

##### Reference
- `let a = 1; value(a);` just echo a, and remove reference
- `let a = 1; echo(a);` just echo a, and with reference (variable)

##### Import / Export
- `import "abc.t";` to get export variable from file abc.t

### Special Usage of Hash
Hash is a special type with abilities to simulate array or function.

A Hash can also be defined as class, or a class instance.

#### Use key from another Hash
##### Define a class
```
let Mark = {
    "@class": "Mark"
};
```
##### Define an instance
```
let m = {
    "@template": Mark
};
```
##### Add a key to the class
```
Mark.greet = func() { printLine "Hello World!"; };
```
##### Use the key in the instance
```
m.greet();
```
This will print "Hello World!"

#### Operator () and \[]
```
Mark["@()"] = func(args) {
    printLine(args[0]);
};
m "Hello World!";
```
This will print "Hello World!"
```
Mark["@[]"] = func(args) {
    printLine(args[0]);
};
m["Hello World!"];
```
This will print "Hello World!"

#### Self parameter
Function with the self parameter will capture its container
```
let a = {};
a.name = "Mark";
a.fn = func(self) {
    printLine("Hello, " + self.name);
};
a.fn();
````
This will print "Hello, Mark"

#### &Self parameter
Function with &self parameter will capture the reference of its container
```
let a = {};
a.name = "Mark";
a.fn = func(&self) {
    &self.xx = "Hello, Mark";
};
printLine(a.xx);
````
This will print "Hello, Mark"
#### Class Type
A hash with "@class" key is a Proto
```
let Mark = {"@class": "Mark"}
printLine(classType Mark);
```
This will print "Proto"


A hash with "@template" key and no "@class" key is an Instance
```
let mark = {"@template": Mark}
printLine(classType mark);
```
This will print "Instance"


A hash with both "@template" key and "@class" key is a Proto (Sub-class)
```
let markChild = {"@template": Mark, "@class": "MarkChild"}
printLine(classType markChild);
```
This will print "Proto"
#### Class Initializer
Standard way to define a class with initializer
Use @() and classType self == "Proto" to initialize class
```
let People = {
    "@class": "People",
    "@()": func(args, self) {
        if (classType self == "Proto") {
            ret {
                "@template": self,
                "name": args[0],
                "age": args[1],
            }
        }
    },
    "greet": func(self) {
        if (classType self == "Instance") {
            printLine("Hi, I am " + self.name)
            printLine("I am " + string(self.age))
        }
    },
}
let mark = People("Mark", 17)
mark.greet()
```
Output:
```
Hi, I am Mark
I am 17
```

Use super(self, name) to access the key on super class
```
let Student = {
    "@class": "Student",
    "@template": People,
    "@()": func(args, self) {
        if (classType self == "Proto") {
            let s = super(self, "@()")(args)
            s.school = args[2]
            ret s
        }
    },
    "greet": func(self) {
        if (classType self == "Instance") {
            super(self, "greet")()
            printLine("I am from " + self.school)
        }
    },
}
let zia = Student("Zia", 16, "CNUHS")
zia.greet()
```
Output:
```
Hi, I am Zia
I am 16
I am from CNUHS
```
