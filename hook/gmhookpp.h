//
// Created by root on 11/20/20.
//

#ifndef GMHOOK_GMHOOKPP_H
#define GMHOOK_GMHOOKPP_H
#include <string>


namespace gomark {
    class GmVariable {
    public:
        GmVariable(){}
        explicit GmVariable(int type, const std::string &name);
        bool expose(int type, const std::string &name);
        ~GmVariable();

        inline bool valid() {
            return var_ != INVALID_VAR_ID;
        }
        GmVariable& operator<<(int32_t value);

    private:
        int type_ = 0;
        int var_ = 0;
    };

}
#endif //GMHOOK_GMHOOKPP_H
