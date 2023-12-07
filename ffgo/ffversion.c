#include "ffgo.h"
#include <libavutil/avutil.h>

void ff_version() {
    av_log(NULL, AV_LOG_INFO, "ffmpeg version: %s\n", av_version_info());
}