cmake_minimum_required(VERSION 3.10)
project(gmhook)

set(CMAKE_CXX_STANDARD 11)

add_library(gmhook SHARED gmhook.cpp gmhook.h)

install(TARGETS gmhook
        RUNTIME DESTINATION bin
        LIBRARY DESTINATION lib
        ARCHIVE DESTINATION lib
        PUBLIC_HEADER DESTINATION include
        )

list(APPEND HDR_FILES
        ${CMAKE_SOURCE_DIR}/gmhook.h
        ${CMAKE_SOURCE_DIR}/gmhookpp.h
        )
install(FILES ${HDR_FILES} DESTINATION include)
