package sdp

var MockSDP = []byte(`v=0
o=- 2251938235 2251938235 IN IP4 0.0.0.0
s=Media Server
c=IN IP4 0.0.0.0
t=0 0
a=control:*
a=packetization-supported:DH
a=rtppayload-supported:DH
a=range:npt=now-
m=video 0 RTP/AVP 96
a=control:trackID=0
a=framerate:25.000000
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1;profile-level-id=64002A;sprop-parameter-sets=Z2QAKqwsaoHgCJ+WbgICAgQA,aO48sAA=
a=recvonly
m=audio 0 RTP/AVP 97
a=control:trackID=1
a=rtpmap:97 MPEG4-GENERIC/16000
a=fmtp:97 streamtype=5;profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;config=1408
a=recvonly
m=audio 0 RTP/AVP 97
a=control:trackID=2
a=rtpmap:97 MPEG4-GENERIC/16000
a=fmtp:97 streamtype=5;profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;config=1408
a=recvonly
`)
