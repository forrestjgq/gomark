package adapter

/*
#include <stdlib.h>
#include <stdint.h>
#include "gmhook.h"

extern int gmCreate(int var_type, const char *name);
extern void gmMark(int id, int value);
extern void gmCancel(int id);
static int gmCreateEx(int var_type, const char *name) {
	return gmCreate(var_type, name);
}
static void initGm() {
	register_gm_hook(gmCreateEx, gmMark, gmCancel);
}
*/
import "C"

func StartAdapter() {
	C.initGm()
}
