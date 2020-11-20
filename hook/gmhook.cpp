#include "gmhook.h"
#include "gmhookpp.h"

namespace gomark {
    class Hook {
    public:
        void Register(gm_var_creator creator, gm_var_marker marker, gm_var_canceler canceler) {
            creator_ = creator;
            marker_ = marker;
            canceler_ = canceler;
        }
        int Create(int var_type, const char *name) {
            if(creator_) {
                return creator_(var_type, name);
            }
            return INVALID_VAR_ID;
        }
        void Mark(int id, int value) {
            if(marker_) {
                marker_(id, value);
            }
        }
        void Cancel(int id) {
            if(canceler_) {
                canceler_(id);
            }
        }
    private:
        gm_var_creator creator_ = nullptr;
        gm_var_marker marker_ = nullptr;
        gm_var_canceler canceler_ = nullptr;
    };

    static Hook hook_;


    GmVariable::GmVariable(int type, const std::string &name) {
        expose(type, name);
    }

    bool GmVariable::expose(int type, const std::string &name) {
        type_ = type;
        var_ = gm_var_create(type, name.c_str());
        return valid();
    }

    GmVariable::~GmVariable() {
        if(valid()) {
            gm_var_cancel(var_);
        }
    }

    GmVariable &GmVariable::operator<<(int32_t value) {
        gm_var_mark(var_, value);
        return *this;
    }
}
void register_gm_hook(gm_var_creator creator, gm_var_marker marker, gm_var_canceler canceler) {
    gomark::hook_.Register(creator, marker, canceler);
}
int gm_var_create(int var_type, const char *name) {
    return gomark::hook_.Create(var_type, name);
}
void gm_var_mark(int id, int value) {
    gomark::hook_.Mark(id, value);
}
void gm_var_cancel(int id) {
    gomark::hook_.Cancel(id);
}
