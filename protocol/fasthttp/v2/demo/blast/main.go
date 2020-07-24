package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

/*
Validation: valid
Context Attributes,
  specversion: 1.0
  type: curl.demo
  source: curl-command
  id: 123-abc
  datacontenttype: application/json
Data,
  {"name": "Dave"}

*/

const data = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec tincidunt non ex vel aliquet. Nunc vestibulum dolor sit amet aliquam commodo. Proin semper est ut tempus placerat. Donec porta viverra metus, ac hendrerit sapien ultrices vitae. Pellentesque hendrerit, tellus pulvinar bibendum tempor, elit enim aliquam nunc, vitae maximus nibh libero at sem. Nullam quam metus, molestie eget euismod at, laoreet ac metus. Ut vitae tortor vestibulum, feugiat nunc vel, rutrum magna. Donec mollis ex consequat, ultricies magna sed, molestie leo. Maecenas sed vestibulum ligula, fringilla scelerisque ligula. Aenean nisl enim, condimentum quis condimentum at, condimentum ac nunc. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed eget est fermentum, maximus nulla nec, suscipit magna. Praesent mattis sapien sed risus semper, eget sagittis arcu efficitur. Pellentesque pharetra varius efficitur.
    Quisque elementum nibh id nisl mattis dapibus. Pellentesque mi sem, aliquam id lacinia tincidunt, consectetur vel turpis. Aenean nec tincidunt urna, non aliquet augue. Curabitur sit amet velit ut arcu congue accumsan eu condimentum elit. Suspendisse id elit a neque varius lobortis. Etiam tempus eros vitae tellus pulvinar posuere. Sed ultricies varius leo, vitae vestibulum quam fringilla nec. Aliquam vitae leo sem. Donec ullamcorper maximus felis, quis tincidunt elit posuere at. Aliquam fringilla pulvinar consectetur. Etiam orci turpis, tincidunt eu metus eget, pharetra blandit turpis. Quisque vestibulum varius mattis. Quisque non posuere augue. Mauris sollicitudin sapien iaculis velit placerat, vel dignissim nisi accumsan. Suspendisse sit amet tortor id nisi placerat eleifend condimentum nec arcu.
    Phasellus condimentum, elit eu pretium condimentum, est diam suscipit dolor, non molestie metus libero quis arcu. Nulla facilisi. Quisque iaculis dignissim fermentum. Nunc rhoncus vitae urna ut maximus. Quisque tempus enim sit amet diam cursus accumsan. Sed condimentum, velit in consequat imperdiet, nulla ipsum efficitur risus, non vehicula urna arcu eget mi. Pellentesque eget sagittis elit. Suspendisse tincidunt mauris a lacus tincidunt ullamcorper. Nulla enim magna, ullamcorper sit amet tristique in, lacinia non enim. Praesent dapibus faucibus metus. Aliquam at ligula quis dui rutrum ullamcorper. Nullam sit amet ligula eu nisi luctus tempor. Aenean finibus ac tortor sed fermentum. Nullam ultrices in ante vel viverra. Aliquam at ligula in nisl iaculis volutpat eget eu lacus.
    In ac nisi imperdiet, egestas leo id, molestie enim. Cras consectetur nisi risus, volutpat sagittis sapien tempor sed. Integer laoreet ligula id mi laoreet, ac auctor tellus sodales. Ut elementum, justo nec accumsan dignissim, dui lacus molestie est, nec imperdiet purus purus sed metus. Vivamus sed felis mollis, pulvinar odio ac, sodales justo. Maecenas pharetra ultricies lorem, nec finibus lacus congue ut. Morbi tincidunt, orci non vehicula interdum, dui elit ornare ex, sit amet bibendum nulla augue a mi. Maecenas vel dictum ante. Nulla vel laoreet lectus, non euismod ex.
    Aenean id tempus ante. Donec fermentum ac quam at placerat. Duis tempor sapien id arcu ornare, in pellentesque ex vehicula. Curabitur at vulputate orci. Maecenas ullamcorper volutpat velit, sit amet sagittis sapien rhoncus non. Maecenas pulvinar, massa eu elementum tincidunt, purus arcu maximus sem, id elementum mi tortor eget turpis. Pellentesque eget est et nibh egestas interdum non vel nunc. Cras facilisis posuere libero vitae vulputate. Proin dignissim eu lectus id tempus. Maecenas tincidunt iaculis ligula, a rutrum nisl blandit vel. Pellentesque in consequat nisi. Praesent fringilla auctor neque, vitae luctus nulla interdum eget. Integer sed dolor vehicula, sollicitudin odio ut, tincidunt purus.
    Donec ac velit id mauris scelerisque laoreet sed a quam. Donec convallis arcu a luctus ornare. Sed ac mi non lorem auctor mattis. Cras maximus nunc sed ipsum egestas, ac imperdiet tortor vulputate. Proin eget nisl erat. Praesent nulla neque, varius id turpis nec, euismod viverra erat. Mauris gravida consectetur pretium. Quisque tincidunt id ante non aliquam. Suspendisse potenti. Donec diam magna, porta eget tempor id, ornare a ligula. Morbi aliquam diam nec tellus aliquam pharetra. Proin consequat nisl non sapien fermentum, non dapibus risus convallis. Donec gravida, orci id pulvinar lobortis, mi arcu suscipit turpis, rhoncus dapibus dui tortor faucibus purus. Quisque blandit vehicula justo id tempus. Donec augue quam, congue a tortor vel, cursus accumsan nibh. Proin rhoncus aliquet ante, ut maximus augue.
    Donec blandit bibendum purus vel rutrum. Fusce odio lacus, bibendum eu ante et, cursus convallis quam. Suspendisse ac lorem et felis auctor hendrerit. Nulla ultricies, libero eget consectetur laoreet, velit eros condimentum dui, vel consequat nisi dolor eget ipsum. Etiam sit amet molestie leo. Maecenas vestibulum lacus a placerat posuere. In sit amet ullamcorper arcu. Nunc enim tellus, lacinia vitae semper sed, finibus in mi. Aliquam egestas dictum suscipit. Curabitur efficitur pulvinar commodo. Pellentesque venenatis sem a massa gravida finibus. Fusce diam eros, accumsan eget aliquet eu, aliquet sed nunc. Cras at massa imperdiet, dignissim lectus et, molestie orci. Vestibulum elementum accumsan tellus, eu fringilla diam dapibus id. In est nisl, fermentum sit amet justo congue, posuere venenatis dolor.
    Ut vestibulum justo id lectus auctor, quis condimentum felis blandit. Aliquam tristique congue aliquam. Duis ornare nulla odio, ut accumsan lacus congue sed. Integer imperdiet dolor ex, iaculis convallis dui ultrices et. Vivamus sollicitudin velit a est malesuada maximus. Nulla ut libero eget ipsum posuere luctus. Morbi id molestie dolor, ut dapibus erat. Pellentesque iaculis fermentum turpis at congue. Cras a leo nec neque dignissim pellentesque viverra sed tortor. Sed orci justo, auctor ut suscipit ultrices, maximus elementum ipsum. Curabitur luctus cursus diam. Duis in efficitur diam. Nunc maximus ullamcorper orci. Donec vitae elementum nisi.
    Vivamus semper ullamcorper orci, a accumsan lacus consectetur ut. Nullam ultricies cursus massa vel tristique. Integer laoreet efficitur diam sed consequat. Ut facilisis dolor dapibus varius accumsan. Donec sed tempus diam, sit amet venenatis justo. Praesent nisi dolor, euismod ut euismod eget, rutrum cursus purus. In semper tortor ac quam lacinia, et cursus ligula ultricies. Phasellus feugiat quam id sagittis sodales. Nullam mattis tortor a porta accumsan. Nam facilisis sagittis nulla quis ultrices. In sed tincidunt est. Donec ac odio et nisi ullamcorper sagittis nec non ante. Aliquam id lorem sed lorem ultrices aliquam vitae et velit.
    Aliquam mattis eros erat, vel malesuada ipsum tempor vitae. Mauris a semper turpis, bibendum auctor ante. Proin euismod orci tellus, vel elementum velit lacinia vitae. Fusce finibus dolor non enim vehicula blandit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nam a convallis ligula, sit amet pellentesque turpis. Phasellus aliquet enim a est pellentesque sollicitudin. Proin a gravida dolor, sed elementum ante. In vitae dolor blandit, gravida justo dignissim, commodo sapien. Sed hendrerit nibh vitae nisl efficitur, ut semper tellus accumsan.
    Sed eget nibh et magna porttitor sagittis. Duis lobortis turpis sit amet rutrum rutrum. Pellentesque massa odio, iaculis gravida iaculis faucibus, scelerisque ut mauris. Duis tempor venenatis nibh, pellentesque lacinia augue elementum finibus. Pellentesque id turpis et risus molestie gravida ut id nisl. Morbi leo eros, lacinia at aliquam et, ultricies sed tellus. Duis lacus tellus, ullamcorper vitae est at, consequat ultrices nisl. Nullam malesuada vehicula dignissim. Nulla facilisi. Etiam quis nibh nibh.
    Quisque ultrices malesuada orci, eget aliquam diam auctor ut. Integer consequat posuere fermentum. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Morbi vestibulum euismod ligula at pulvinar. Nam magna lectus, imperdiet varius congue finibus, lobortis sed nibh. Aliquam aliquam ante eget justo egestas, vitae sollicitudin orci auctor. Vivamus dictum tincidunt placerat. Nam fringilla vel quam ut porttitor. Vestibulum dictum elit a hendrerit elementum. Curabitur vitae neque vitae metus mattis viverra in nec risus. Maecenas hendrerit velit at malesuada pharetra.
    Phasellus ac ligula urna. Phasellus nec mi purus. Maecenas lacinia turpis sed urna commodo lacinia. Phasellus posuere imperdiet eros, vel iaculis purus ornare eget. Donec quis placerat elit. Morbi congue eros a justo pretium, in semper nisl ultricies. Nunc felis velit, maximus non nisi ut, convallis vulputate ex. Vivamus aliquet sodales finibus. In vestibulum ligula eget laoreet interdum. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus.
    Vestibulum efficitur massa sapien, ut ornare mi finibus vel. Quisque molestie est sit amet risus posuere, eu eleifend velit tempus. Pellentesque iaculis risus nulla, ut malesuada odio vehicula vel. Phasellus molestie semper leo, sed vehicula libero mollis nec. In hac habitasse platea dictumst. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Fusce porta, elit ut dignissim eleifend, libero elit aliquam erat, ut commodo urna sem et tortor. Donec cursus, sem vel dapibus fringilla, lacus lectus facilisis magna, eu volutpat turpis felis non orci. Quisque at rutrum dui. Donec mattis tristique elementum. Phasellus sollicitudin a arcu vel tincidunt. Integer pellentesque lacus risus, ut ullamcorper nisl mattis ut. Proin in mi nec lacus hendrerit consequat sed ut nibh. Aenean interdum quam tellus, eu vestibulum nibh aliquet eu. Nulla aliquet a enim facilisis posuere.
    Nunc dapibus aliquet risus sed aliquet. Quisque finibus orci diam, eget rhoncus magna ultricies nec. Nam eu magna at neque feugiat pellentesque. Sed ut felis in metus ultricies interdum sed vitae enim. Sed laoreet.`

func main() {
	b := fmt.Sprintf("{%q:%q}", "lorem", data)

	c := fasthttp.Client{}
	for {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("http://localhost:8080")
		req.Header.SetMethod(fasthttp.MethodPost)
		req.Header.Set("Ce-Specversion", "1.0")
		req.Header.Set("Ce-Type", "fasthttp.blast")
		req.Header.Set("Ce-Source", "fasthttp/blast")
		req.Header.Set("Ce-Id", "abc-123")
		req.Header.SetContentType("application/json")
		req.SetBodyString(b)

		resp := fasthttp.AcquireResponse()

		if err := c.Do(req, resp); err != nil {
			panic(err)
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			panic(fmt.Sprintf("unexpected status code %d. Expecting %d", resp.StatusCode(), fasthttp.StatusOK))
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}
}
