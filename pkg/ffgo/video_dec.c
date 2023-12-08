// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
// #include <setjmp.h>
// #include <string.h>
// #include <errno.h>

// #include "lzy_log.h"
// #include "Pub/lzy_type.h"

// #include "libavcodec/avcodec.h"
// #include "libavformat/avformat.h"
// #include "libswscale/swscale.h"
// #include "libavutil/frame.h"


// #ifdef __cplusplus
// #if __cplusplus
// extern "C" {
// #endif
// #endif /* End of #ifdef __cplusplus */

// /*
//  *
//  */
// static void SaveFrame(char *pDataRGB, LZY_S32 width, LZY_S32 height, LZY_S32 iFrame)
// {
//     FILE *pFile = NULL;
//     char szFilename[32] = {0};

//     if(iFrame > 5)
//     {
//         return;
//     }

//     sprintf(szFilename, "/tmp/mnt/sdcard/%d.ppm", iFrame);
//     pFile = fopen(szFilename, "wb");
//     if(NULL == pFile)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "fopen(%s) error\n", szFilename);
//         return;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "iFrame:%d\n", iFrame);

//     //fprintf(pFile, "P6\n%d %d\n255\n", width, height);
//     //fwrite(pDataRGB, 1, height * width * 3, pFile);
//     fwrite(pDataRGB, 1, height * width * 2, pFile);

//     fclose(pFile);
// }

// /*
//  *
//  */
// LZY_S32 video_dec_Mp4ToRGB565()
// {
//     char filepath[] = "/tmp/mnt/sdcard/MD_012.mp4";
//     LZY_S32 nCount = 0;
//     LZY_S32 i = 0, videoIndex = -1, numBytes = 0, numBytesRGB = 0;
//     LZY_S32 outLinesizeRGB[4] = {0};
//     LZY_S32 ret = 0, ret1 = 0;
//     LZY_U8 *outBufferRGB = NULL;
//     LZY_U8 *outDataRGB[4] = {NULL, NULL, NULL, NULL};

//     AVFormatContext *pFormatCtx = NULL;
//     AVCodecContext *pCodecCtx = NULL;
//     AVCodec *pCodec = NULL;
//     AVFrame *pFrame = NULL, *pFrameYUV = NULL;
//     AVPacket packet;
//     struct SwsContext *img_convert_ctx = NULL;

//     pFormatCtx = avformat_alloc_context();
//     if(NULL == pFormatCtx)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "avformat_alloc_context error\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avformat_open_input:%s\n", filepath);

//     ret = avformat_open_input(&pFormatCtx, filepath, NULL, NULL);
//     if(ret != 0)
//     {
//         char *msg = av_err2str(ret);
//         LZY_LOG(LZY_LOG_ERROR, "Can't open the file(%s), ret:%#x, %s\n", filepath, ret, msg);
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avformat_find_stream_info\n");

//     ret = avformat_find_stream_info(pFormatCtx, NULL);
//     if(ret < 0)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "Couldn't find stream information.\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "av_dump_format\n");

//     av_dump_format(pFormatCtx, 0, filepath, 0);

//     for(i = 0; i < pFormatCtx->nb_streams; i++)
//     {
//         if(pFormatCtx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO)
//         {
//             videoIndex = i;
//             break;
//         }
//     }

//     if(videoIndex == -1)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "videoIndex error.\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avcodec_find_decoder\n");

//     pCodec = avcodec_find_decoder(pFormatCtx->streams[videoIndex]->codecpar->codec_id);
//     if(NULL == pCodec)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "Unsupported codec!\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avcodec_alloc_context3\n");

//     pCodecCtx = avcodec_alloc_context3(pCodec);
//     if(NULL == pCodecCtx)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "avcodec_alloc_context3 error\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avcodec_parameters_to_context\n");

//     ret = avcodec_parameters_to_context(pCodecCtx, pFormatCtx->streams[videoIndex]->codecpar);
//     if(ret < 0)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "Could not copy parameters to codec.\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "avcodec_open2\n");

//     ret = avcodec_open2(pCodecCtx, pCodec, NULL);
//     if(ret < 0)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "Could not open codec.\n");
//         return -1;
//     }

//     pFrame = av_frame_alloc();
//     if(NULL == pFrame)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "av_frame_alloc error\n");
//         return -1;
//     }

//     pFrameYUV = av_frame_alloc();
//     if(NULL == pFrameYUV)
//     {
//         LZY_LOG(LZY_LOG_ERROR, "av_frame_alloc error\n");
//         return -1;
//     }

//     LZY_LOG(LZY_LOG_PRINT, "sws_getContext\n");

//     img_convert_ctx = sws_getContext(pCodecCtx->width, pCodecCtx->height, pCodecCtx->pix_fmt,\
//                                         pCodecCtx->width, pCodecCtx->height, AV_PIX_FMT_RGB565LE, SWS_POINT, NULL, NULL, NULL);

//     LZY_LOG(LZY_LOG_PRINT, "av_image_get_buffer_size\n");

//     numBytesRGB = av_image_get_buffer_size(AV_PIX_FMT_RGB565LE, pCodecCtx->width, pCodecCtx->height, 1);
//     outBufferRGB = (uint8_t*)av_malloc(numBytesRGB * sizeof(uint8_t));
 
//     outDataRGB[0] = outBufferRGB;
//     //outLinesizeRGB[0] = pCodecCtx->width * 3;
//     outLinesizeRGB[0] = pCodecCtx->width * 2;
//     outLinesizeRGB[1] = outLinesizeRGB[2] = outLinesizeRGB[3] = 0;
//     av_new_packet(&packet, numBytes);

//     while(av_read_frame(pFormatCtx, &packet) >= 0)
//     {
//         if(packet.stream_index == videoIndex) 
//         {
//             //ret = avcodec_decode_video2(pCodecCtx, pFrame, &frameFinished, &packet);
//             //if (ret >= 0) 
//             LZY_LOG(LZY_LOG_PRINT, "avcodec_send_packet\n");
//             if((ret = avcodec_send_packet(pCodecCtx, &packet)) >= 0  && (ret1 = avcodec_receive_frame(pCodecCtx, pFrame)) >= 0)
//             {
//                 LZY_LOG(LZY_LOG_PRINT, "avcodec_receive_frame\n");

//                 sws_scale(img_convert_ctx, (const uint8_t* const*)pFrame->data,pFrame->linesize, 0,pCodecCtx->height, outDataRGB, outLinesizeRGB);

//                 LZY_LOG(LZY_LOG_PRINT, "sws_scale\n");

//                 VideoPlayback(outBufferRGB);

//                 //SaveFrame(outBufferRGB,  pCodecCtx->width, pCodecCtx->height, nCount);
//                 ++nCount;

//                 LZY_LOG(LZY_LOG_PRINT, "SaveFrame\n");
//             }
//         }

//         av_packet_unref(&packet);
//     }
 
//     av_free(outBufferRGB);
//     av_frame_free(&pFrame);
//     av_frame_free(&pFrameYUV);
//     avcodec_close(pCodecCtx);
//     avformat_close_input(&pFormatCtx);
 
//     return 0;
// }



// #ifdef __cplusplus
// #if __cplusplus
// }
// #endif
// #endif /* End of #ifdef __cplusplus */
