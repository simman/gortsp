#include "ffgo.h"
//#include <stdio.h>
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>

int x265_to_h264() {
    av_log_set_level(AV_LOG_VERBOSE);
    avformat_network_init();

    AVFormatContext *inputContext = NULL;
    AVFormatContext *outputContext = NULL;
    AVStream *inputStream = NULL;
    AVStream *outputStream = NULL;
    AVCodecContext *inputCodecContext = NULL;
    AVCodecContext *outputCodecContext = NULL;
    AVPacket packet;
    int ret;

    const char *inputFile = "http://192.168.5.222:8666/video_full.h265";
    
#ifdef IS_IOS
    NSArray *paths = NSSearchPathForDirectoriesInDomains(NSDocumentDirectory, NSUserDomainMask, YES);
    NSString *docPath = [paths lastObject];
    NSString *filePath = [NSString stringWithFormat:@"%@/output.h264", docPath];
    const char *outputFile = [filePath UTF8String];
#else
    const char *outputFile = "output.h264";
#endif

    // 打开输入文件
    ret = avformat_open_input(&inputContext, inputFile, NULL, NULL);
    if (ret < 0) {
        printf("无法打开输入文件\n");
        return ret;
    }

    // 获取输入流信息
    ret = avformat_find_stream_info(inputContext, NULL);
    if (ret < 0) {
        printf("无法获取输入流信息\n");
        return ret;
    }
    
    av_dump_format(inputContext, 0, inputFile, 0);

    // 查找视频流
    ret = av_find_best_stream(inputContext, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (ret < 0) {
        printf("无法找到视频流\n");
        return ret;
    }

    // 获取视频流
    inputStream = inputContext->streams[ret];
    inputCodecContext = inputStream->codec;
    
    // 查找解码器
    AVCodec *inputCodec = avcodec_find_decoder(AV_CODEC_ID_H265);
    if (!inputCodec) {
        printf("无法找到H.265解码器\n");
        return AVERROR_UNKNOWN;
    }
    
    // 打开解码器
    ret = avcodec_open2(inputCodecContext, inputCodec, NULL);
    if (ret < 0) {
        printf("无法打开编码器\n");
        return ret;
    }
    

    // 初始化输出格式上下文
    avformat_alloc_output_context2(&outputContext, NULL, NULL, outputFile);
    if (!outputContext) {
        printf("无法创建输出格式上下文\n");
        return AVERROR_UNKNOWN;
    }

    // 创建输出流
    outputStream = avformat_new_stream(outputContext, NULL);
    if (!outputStream) {
        printf("无法创建输出流\n");
        return AVERROR_UNKNOWN;
    }

    // 复制输入流的参数到输出流
    ret = avcodec_parameters_copy(outputStream->codecpar, inputStream->codecpar);
    if (ret < 0) {
        printf("无法复制参数\n");
        return ret;
    }

    // 查找编码器
    AVCodec *outputCodec = avcodec_find_encoder(AV_CODEC_ID_H264);
    if (!outputCodec) {
        printf("无法找到H.264编码器\n");
        return AVERROR_UNKNOWN;
    }

    // 初始化编码器上下文
    outputCodecContext = avcodec_alloc_context3(outputCodec);
    if (!outputCodecContext) {
        printf("无法创建编码器上下文\n");
        return AVERROR_UNKNOWN;
    }

    // 设置编码器参数
    outputCodecContext->codec_id = AV_CODEC_ID_H264;
    outputCodecContext->codec_type = AVMEDIA_TYPE_VIDEO;
    outputCodecContext->width = inputCodecContext->width;
    outputCodecContext->height = inputCodecContext->height;
    outputCodecContext->pix_fmt = AV_PIX_FMT_YUV420P; //outputCodec->pix_fmts[0];
    outputCodecContext->time_base = inputStream->time_base;

    // 打开编码器
    ret = avcodec_open2(outputCodecContext, outputCodec, NULL);
    if (ret < 0) {
        printf("无法打开编码器\n");
        return ret;
    }

    // 打开输出文件
    ret = avio_open(&outputContext->pb, outputFile, AVIO_FLAG_WRITE);
    if (ret < 0) {
        printf("无法打开输出文件\n");
        return ret;
    }

    // 写入输出文件的头部
    ret = avformat_write_header(outputContext, NULL);
    if (ret < 0) {
        printf("无法写入输出文件头部\n");
        return ret;
    }

    // 初始化packet
    av_init_packet(&packet);
    packet.data = NULL;
    packet.size = 0;

    int frameCount = 0;
    // 读取输入文件的packet并进行转换
    while (av_read_frame(inputContext, &packet) >= 0) {
        if (packet.stream_index == inputStream->index) {
            // 发送packet到解码器
            ret = avcodec_send_packet(inputCodecContext, &packet);
            if (ret < 0) {
                printf("无法发送packet到解码器\n");
                if (ret == AVERROR(EINVAL)) {
                    printf("找不到编码器");
                }
                break;
            }

            // 从解码器接收解码后的frame
            AVFrame *frame = av_frame_alloc();
            frameCount++;
            while (ret >= 0) {
                ret = avcodec_receive_frame(inputCodecContext, frame);
                if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
                    break;
                } else if (ret < 0) {
                    printf("无法从解码器接收帧\n");
                    break;
                }

                // 发送frame到编码器
                ret = avcodec_send_frame(outputCodecContext, frame);
                if (ret < 0) {
                    printf("无法发送帧到编码器\n");
                    break;
                }

                // 从编码器接收编码后的packet并写入输出文件
                while (ret >= 0) {
                    ret = avcodec_receive_packet(outputCodecContext, &packet);
                    if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
                        break;
                    } else if (ret < 0) {
                        printf("无法从编码器接收packet\n");
                        break;
                    }

                    // 写入packet到输出文件
                    ret = av_write_frame(outputContext, &packet);
                    if (ret < 0) {
                        printf("无法写入packet到输出文件\n");
                        break;
                    }

                    av_packet_unref(&packet);
                }

                av_frame_unref(frame);
            }

            av_frame_free(&frame);
        }

        av_packet_unref(&packet);
    }

    printf("frameCount: \n");
    printf("%d", frameCount);
    // 写入输出文件的尾部
    ret = av_write_trailer(outputContext);
    if (ret < 0) {
        printf("无法写入输出文件尾部\n");
        return ret;
    } else {
        printf("转换完成!!!!!");
    }

    // 释放资源
    avcodec_free_context(&inputCodecContext);
    avcodec_free_context(&outputCodecContext);
    if (inputContext != NULL) {
//        avformat_close_input(&inputContext);
    }
    avio_closep(&outputContext->pb);
    avformat_free_context(outputContext);
    
    uint8_t *outputBuffer = NULL;
        int outputBufferSize;

    return 10086;
}