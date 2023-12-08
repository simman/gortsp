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
    ff->pInputVideoCodec = NULL;
    ff->pInputAudioCodecCtx = NULL;
    ff->pInputAudioCodec = NULL;
    ff->packet = NULL;
    ff->pFrame = NULL;

    return ff;
}

int videoStream = -1;

int create_decoder(ff_rtsp_client *ff) {
    // Find the first video stream
    printf("调用 - av_find_best_stream\n");
    videoStream = av_find_best_stream(ff->pInputFormatCtx, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (videoStream == -1) {
        return AVERROR_UNKNOWN;
    }

    printf("调用 - avcodec_find_decoder\n");
    // Find the decoder for the video stream
    ff->pInputVideoCodec = avcodec_find_decoder(ff->pInputFormatCtx->streams[videoStream]->codecpar->codec_id);
    if (ff->pInputVideoCodec == NULL) {
        av_log(NULL, AV_LOG_ERROR, "Unsupported codec!\n");
        return AVERROR_UNKNOWN;
    }

    printf("调用 - avcodec_alloc_context3\n");
    // Get a pointer to the codec context for the video stream
    ff->pInputVideoCodecCtx = avcodec_alloc_context3(ff->pInputVideoCodec);
    
    // 获取输入视频流
    ff->pInputVideoStream = ff->pInputFormatCtx->streams[videoStream];

    int ret = avcodec_parameters_to_context(ff->pInputVideoCodecCtx, ff->pInputFormatCtx->streams[videoStream]->codecpar);
    if(ret < 0) {
        av_log(NULL, AV_LOG_ERROR, "Could not copy parameters to codec.\n");
        return AVERROR_UNKNOWN;
    }

    printf("调用 - avcodec_open2\n");
    // Open codec
    if (avcodec_open2(ff->pInputVideoCodecCtx, ff->pInputVideoCodec, NULL) < 0) {
        av_log(NULL, AV_LOG_ERROR, "Cannot open video decoder\n");
        return AVERROR_UNKNOWN;
    }

    return 0;
}

int create_encoder(ff_rtsp_client *ff) {
    // 查找编码器
    ff->pOutputVideoCodec = avcodec_find_encoder(AV_CODEC_ID_H264);
    if (ff->pOutputVideoCodec == NULL) {
        printf("无法找到H.264编码器\n");
        return AVERROR_UNKNOWN;
    }

    printf("调用 - avcodec_alloc_context3\n");
    ff->pOutputVideoCodecCtx = avcodec_alloc_context3(ff->pOutputVideoCodec);
    if (ff->pOutputVideoCodecCtx == NULL) {
        printf("无法创建编码器上下文\n");
        return AVERROR_UNKNOWN;
    }

    // 设置编码器参数
    ff->pOutputVideoCodecCtx->codec_id = AV_CODEC_ID_H264;
    ff->pOutputVideoCodecCtx->codec_type = AVMEDIA_TYPE_VIDEO;
    ff->pOutputVideoCodecCtx->width = ff->pInputVideoCodecCtx->width;
    ff->pOutputVideoCodecCtx->height = ff->pInputVideoCodecCtx->height;
    ff->pOutputVideoCodecCtx->pix_fmt = AV_PIX_FMT_YUV420P; //outputCodec->pix_fmts[0];
    ff->pOutputVideoCodecCtx->time_base = ff->pInputVideoStream->time_base;
    
    printf("调用 - avcodec_open2\n");
    if (avcodec_open2(ff->pOutputVideoCodecCtx, ff->pOutputVideoCodec, NULL) < 0) {
        printf("无法打开编码器\n");
        return AVERROR_UNKNOWN;
    }

    return 0;
}

int prepare_rtsp_client(ff_rtsp_client *ff, char *url) {
    av_log_set_level(AV_LOG_VERBOSE);
    avformat_network_init();

    ff->url = url;

    AVDictionary *opts = 0;
    
//    av_dict_set(&opts, "probesize", "6048000", 0);
    printf("调用 - avformat_open_input\n");
    if (avformat_open_input(&ff->pInputFormatCtx, url, NULL, &opts) !=0 ) {
        av_log(NULL, AV_LOG_ERROR, "Couldn't open file\n");
        goto initError;
    }

    printf("调用 - avformat_find_stream_info\n");
    if (avformat_find_stream_info(ff->pInputFormatCtx, NULL) < 0) {
        av_log(NULL, AV_LOG_ERROR, "Couldn't find stream information\n");
        goto initError;
    }

     // 创建解码器
    printf("创建解码器");
    if (create_decoder(ff) != 0) {
        goto initError;
    }

    printf("av_dump_format");
    av_dump_format(ff->pInputFormatCtx, 0, url, 0);


    const char *outputFile = "output.h264";
    // 初始化输出格式上下文
    avformat_alloc_output_context2(&ff->pOutputFormatCtx, NULL, NULL, outputFile);
    if (ff->pOutputFormatCtx == NULL) {
        printf("无法创建输出格式上下文\n");
        return AVERROR_UNKNOWN;
    }

    // 创建编码器
    printf("创建编码器");
    if (create_encoder(ff) != 0) {
        goto initError;
    }

    AVPacket packet;
    av_init_packet(&packet);
    packet.data = NULL;
    packet.size = 0;

    AVFrame * pFrame = av_frame_alloc();

    
    printf("调用 - av_read_frame\n");
    av_log(NULL, AV_LOG_ERROR, "开始解码\n");
    while (av_read_frame(ff->pInputFormatCtx, &packet) >= 0) {
        if (packet.stream_index != videoStream) {
            continue;
        }

        if (packet.size <= 0) {
            av_log(NULL, AV_LOG_ERROR, "跳过空包\n");
            continue;
        }
        
        int ret = avcodec_send_packet(ff->pInputVideoCodecCtx, &packet);
        if (ret < 0 || ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
            av_log(NULL, AV_LOG_ERROR, "无法发送packet到解码器\n");
            continue;
        }

        while (ret >= 0) {
            int result = avcodec_receive_frame(ff->pInputVideoCodecCtx, pFrame);
            if (result == AVERROR(EAGAIN) || result == AVERROR_EOF) {
                break;
            }
            // ----------------- 开始转码 ---------------------
            // 发送frame到编码器
             ret = avcodec_send_frame(ff->pOutputVideoCodecCtx, pFrame);
             if (ret < 0) {
                 printf("无法发送帧到编码器\n");
                 break;
             }

             // 从编码器接收编码后的packet并写入输出文件
             while (ret >= 0) {
                 ret = avcodec_receive_packet(ff->pOutputVideoCodecCtx, &packet);
                 if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
                     break;
                 } else if (ret < 0) {
                     printf("无法从编码器接收packet\n");
                     break;
                 }
                 
                 // 写入packet到输出文件
                 printf("写入packet到输出文件\n");
                 av_packet_unref(&packet);
             }

             av_frame_unref(pFrame);
        }

        av_packet_unref(&packet);
    }
    av_log(NULL, AV_LOG_ERROR, "EXIT ----->\n");

    return 0;
initError:
    return -1;
}

void uninit_rtsp_client(ff_rtsp_client *ff) {
    avformat_close_input(ff->pInputFormatCtx);

    avcodec_free_context(ff->pInputVideoCodecCtx);
    avcodec_free_context(ff->pInputAudioCodecCtx);

    avcodec_close(ff->pInputVideoCodecCtx);
    avcodec_close(ff->pInputAudioCodecCtx);

    av_packet_free(ff->packet);
    av_frame_free(ff->pFrame);

    free(ff->url);
    free(ff);
}