cmake_minimum_required(VERSION 3.10)
project(gmhook)

set(CMAKE_CXX_STANDARD 11)

list(APPEND EXLIBS "")

# BRPC is path of brpc installation
if (BRPC)
    add_definitions(-DBRPC)
    include_directories(${BRPC}/include)
    list(APPEND EXLIBS "-lbrpc")
    link_directories(${BRPC}/lib)
endif ()

add_library(gmhook SHARED gmhook.cpp gmhook.h)
target_link_libraries(gmhook "-Wl,-Bdynamic" ${EXLIBS})

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
