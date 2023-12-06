#include "ffgo.h"
//#include <stdio.h>
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavutil/avutil.h>

ff_rtsp_client* init_rtsp_client() {
    av_log_set_level(AV_LOG_VERBOSE);
    avformat_network_init();

    ff_rtsp_client *ff = (ff_rtsp_client*)malloc(sizeof(ff_rtsp_client));
    ff->url = NULL;
    ff->pInputFormatCtx = NULL;
    ff->pInputVideoStream = NULL;
    ff->pInputAudioStream = NULL;
    ff->pInputVideoCodecCtx = NULL;
    ff->pInputAudioCodecCtx = NULL;
    ff->pInputVideoCodec = NULL;
    ff->pInputAudioCodec = NULL;
    ff->packet = NULL;
    ff->pFrame = NULL;

    return ff;
}

int prepare_rtsp_client(ff_rtsp_client *ff, char *url) {
    av_log_set_level(AV_LOG_VERBOSE);
    avformat_network_init();

    ff->url = url;

    AVDictionary *opts = 0;
    
//    av_dict_set(&opts, "probesize", "6048000", 0);

    if (avformat_open_input(&ff->pInputFormatCtx, url, NULL, &opts) !=0 ) {
        av_log(NULL, AV_LOG_ERROR, "Couldn't open file\n");
        goto initError;
    }

    if (avformat_find_stream_info(&ff->pInputFormatCtx, NULL) < 0) {
        av_log(NULL, AV_LOG_ERROR, "Couldn't find stream information\n");
        goto initError;
    }

    // Find the first video stream
    int videoStream = av_find_best_stream(ff->pInputFormatCtx, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (videoStream == -1) {
        goto initError;
    }
    
    // Find the decoder for the video stream
    ff->pInputVideoCodec = avcodec_find_decoder(ff->pInputFormatCtx->streams[videoStream]->codecpar->codec_id);
    if (ff->pInputVideoCodec == NULL) {
        av_log(NULL, AV_LOG_ERROR, "Unsupported codec!\n");
        goto initError;
    }

    // Get a pointer to the codec context for the video stream
    ff->pInputVideoCodecCtx = avcodec_alloc_context3(ff->pInputVideoCodec);
    
    // Open codec
    if (avcodec_open2(ff->pInputVideoCodecCtx, ff->pInputVideoCodec, NULL) < 0) {
        av_log(NULL, AV_LOG_ERROR, "Cannot open video decoder\n");
        goto initError;
    }

    // Allocate video frame
    // pFrame = av_frame_alloc();
    return 0;
initError:
    return -1;
}

void uninit_rtsp_client(ff_rtsp_client *ff) {
    avformat_close_input(&ff->pInputFormatCtx);

    avcodec_free_context(&ff->pInputVideoCodecCtx);
    avcodec_free_context(&ff->pInputAudioCodecCtx);

    avcodec_close(ff->pInputVideoCodecCtx);
    avcodec_close(ff->pInputAudioCodecCtx);

    av_packet_free(&ff->packet);
    av_frame_free(&ff->pFrame);

    free(ff->url);
    free(ff);
}