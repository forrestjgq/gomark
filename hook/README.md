# What is this?
`gomark` is designed to be a replacement of bvar, a component of brpc, in GO. It works and servers as HTTP independantly.

But sometime GO must work with c/cpp programs, in which marking is required. In old days, those c/cpp programs use bvar independantly, but when they merge with GO program as dynamic library, things changes.

Firstly, brpc can not be linked by GO cgo due to `pthread_mutex_lock` symbol selection issue, which is commented inside bvar's code and is not solved. This makes any c/cpp program using bvar can not be used by GO directly by cgo.

Secondly, both gomark and bvar are combined by their own HTTP service, so you need open seperate web pages to monitor them. It's not right.

So gomark hook is presented. It provides a unified interface to connect to gomark or bvar, acting as a bridge to these services. By compiling with different option, program can mark through gomark hook to either bvar or gomark. Amazing thing is, this selection is transparent to program.

# Compile

To compile a gomark hook to work with gomark, just:
```sh
mkdir -p build
cd build
cmake -DCMAKE_INSTALL_PREFIX=/path/to/install/directory .. && make && make install 
```

To compile a gomark hook to work with bvar, you need first compile brpc and install them to a directory which  has subdirectory `include` and `lib` taking header files and `libbrpc.so`. Then build:
```sh
mkdir -p build
cd build
cmake -DBRPC=/path/to/brpc -DCMAKE_INSTALL_PREFIX=/path/to/install/directory .. && make && make install 
```

gomark hook will install:
- `gmhook.h`: c interface of gmhook
- `gmhookpp.h`: c++ interface of gmhook
- `libgmhook.so`: gmhook library 

# Usage
## Vairable types
As defined in gmhook.h:
```c
// gmhook.h
#define VAR_LATENCY_RECORDER 0
#define VAR_ADDER 1
#define VAR_MAXER 2
#define VAR_STATUS 3
#define VAR_PERSECOND_ADDER 4
```
Each of these defines a variable type.

## C user
C program should call these functions to create/mark/cancel variable:
```c
// gmhook.h
int gm_var_create(int var_type, const char *name);
void gm_var_mark(int id, int value);
void gm_var_cancel(int id);
```

## (CPP)GmVariable and how to use

gomark hook provide a unified adapter:
```cpp
// gmhookpp.h
class GmVariable {
public:
    GmVariable(){}
    explicit GmVariable(int type, const std::string &name);
    bool expose(int type, const std::string &name);
    ~GmVariable();

    bool valid();
    GmVariable& operator<<(int32_t value);
};
```

A `GmVariable` acts as any kind of variable either bvar or gomark provides. To specify variable type, you need to provide a `type` in constructor or `expose()` along with variable `name`. 

By calling `<<` you may mark a value, only this value is an `int32_t`, but it should be large enough.

Here is an example:

```cpp
GmVariable recorder(VAR_LATENCY_RECORDER, "recorder");
recoder << 4;
recorder << val1 << val2 << val3;
// when recorder destructs, it dispose resources automatically.
```
## (CPP)Work with bvar
You may familiar with bvar usage:
```cpp
bvar::LatencyRecorder recorder("some_recorder");
recoder << 4;
recorder << val1 << val2 << val3;
// when recorder destructs, it dispose resources automatically.
```

You can see that it differs from bvar only at declaration(or exposing).

If your program is using bvar now, and you need to work with gomark and brpc selectively, you need just change your variable declaration, and it will work.

## Work with gomark
gomark hook actually act as an bridge to gomark, but it is not gomark. So if you need gmhook to work with gomark, your GO program must load gomark and initialize it by calling:
```c
// gmhook.h
typedef int (*gm_var_creator)(int var_type, const char *name);
typedef void (*gm_var_marker)(int id, int value);
typedef void (*gm_var_canceler)(int id);

void register_gm_hook(gm_var_creator creator, gm_var_marker marker, gm_var_canceler deleter);
```

By using cgo, it injects three callback functions to do variable creation, marking, and destruction.

How gomark injects them is outside of this document.

Interesting thing is, if this register is not called, gmhook will not report any error, just do nothing. This feature is great because it let gmhook user can run without gomark like running a unit test...

