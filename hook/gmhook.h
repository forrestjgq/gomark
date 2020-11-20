#ifndef GMHOOK_GMHOOK_H
#define GMHOOK_GMHOOK_H

#ifdef __cplusplus
extern "C" {
#endif

#define INVALID_VAR_ID 0

#define VAR_LATENCY_RECORDER 0
#define VAR_ADDER 1
#define VAR_MAXER 2
#define VAR_STATUS 3
#define VAR_PERSECOND_ADDER 4

typedef int (*gm_var_creator)(int var_type, const char *name);
typedef void (*gm_var_marker)(int id, int value);
typedef void (*gm_var_canceler)(int id);

void register_gm_hook(gm_var_creator creator, gm_var_marker marker, gm_var_canceler deleter);

int gm_var_create(int var_type, const char *name);
void gm_var_mark(int id, int value);
void gm_var_cancel(int id);

#ifdef __cplusplus
};
#endif

#endif //GMHOOK_GMHOOK_H
