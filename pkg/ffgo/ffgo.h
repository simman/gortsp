#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavutil/avutil.h>

typedef struct {
    char * url;

    AVFormatContext *pInputFormatCtx;
    AVStream *pInputVideoStream;
    AVStream *pInputAudioStream;
    AVCodecContext *pInputVideoCodecCtx;
    AVCodecContext *pInputAudioCodecCtx;
    AVCodec *pInputVideoCodec;
    AVCodec *pInputAudioCodec;
    AVPacket *packet;
    AVFrame *pFrame;

    AVFormatContext *pOutputFormatCtx;
    AVStream *pOutputVideoStream;
    AVStream *pOutputAudioStream;
    AVCodecContext *pOutputVideoCodecCtx;
    AVCodecContext *pOutputAudioCodecCtx;
    AVCodec *pOutputVideoCodec;
    AVCodec *pOutputAudioCodec;
    AVPacket *pOutputPacket;
    AVFrame *pOutputFrame;
} ff_rtsp_client;

int      x265_to_h264();
ff_rtsp_client* init_rtsp_client();
int     prepare_rtsp_client(ff_rtsp_client* c, char *url);
void    uninit_rtsp_client(ff_rtsp_client* c);
void    ff_version();
