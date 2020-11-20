//
// Created by root on 11/20/20.
//

#ifndef GMHOOK_GMHOOKPP_H
#define GMHOOK_GMHOOKPP_H
#include <string>
#include <functional>


namespace gomark {

    class GmVariable {
    public:
        GmVariable(){}
        explicit GmVariable(int type, const std::string &name);
        bool expose(int type, const std::string &name);
        ~GmVariable();

        bool valid();
        GmVariable& operator<<(int32_t value);

    private:
        int type_ = 0;
#ifdef BRPC
        using Markable = std::function<void(int32_t)>;
        Markable markable_;
#else
        int var_ = 0;
#endif
    };

}
#endif //GMHOOK_GMHOOKPP_H
