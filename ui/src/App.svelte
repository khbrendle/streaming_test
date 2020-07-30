<script>
			import vis from "vis-timeline/dist/vis-timeline-graph2d.min";
			import { onMount, onDestroy } from "svelte";

			let stream = "customer_count";
			let container;
			let graph2d;

			var items = [];
			var dataset = new vis.DataSet(items);

			var start = new Date();
			console.log("start: ", start);
			var end = new Date(new Date().setHours(start.getHours() + 2));
			console.log("end: ", end);
			var options = {
			  start: start,
			  end: end,
			  drawPoints: {
			    size: 1,
			    style: "circle"
			  }
			};

			let dataStream;
			let streamHistory = true;
			let groupMinute = 1;

			onMount(() => {
			  dataStream = new EventSource(
			    `http://localhost:3000/v0/stream/subscribe/${stream}?groupMinute=${groupMinute}`
			  );
			  dataStream.onmessage = function(event) {
			    var dat = JSON.parse(event.data);

			    if (streamHistory) {
			      console.log("inital data: ", dat);
			      dat = dat.map(v => {
			        v.x = new Date(v.x).setMilliseconds(0);
			        return v;
			      });
			      dataset.add(dat);
			      streamHistory = false;
			      return;
			    }

			    var posX = new Date(dat.timestamp);
			    posX = new Date(new Date(posX.setSeconds(0)).setMilliseconds(0));
			    var min = posX.getMinutes();
			    min = Math.floor(min / groupMinute) * groupMinute;
			    var posXn = posX.setMinutes(min);
			    console.log("date: ", posXn);

			    var val = dataset.get(
			      {
			        filter: function(item) {
			          // console.log("filter item: ", item);
			          return item.x == posXn;
			        }
			      },
			      {
			        type: {
			          y: "Number"
			        }
			      }
			    );
			    // console.log("filter got: ", val);
			    if (val.length === 0) {
			      // does not exist, create
			      // console.log("adding new value");
			      val = {
			        x: posXn,
			        y: dat.n
			      };
			      // console.log("new value: ", val);
			      dataset.add(val);
			    } else {
			      val = val[0];
			      // exists, update
			      // console.log("updating existing value");
			      // console.log("current value: ", val);
			      val.y += dat.n;
			      // console.log("new value: ", val);
			      dataset.update(val);
			    }
			  };
			  dataStream.onerror = function(err) {
			    console.log("stream error: ", err);
			  };

			  graph2d = new vis.Graph2d(container, dataset, options);
			});

			onDestroy(() => {
			  dataStream.close();
			});
</script>
<svelte:head>
	<link rel="stylesheet" href="/vis-timeline-graph2d.min.css">
</svelte:head>

<h1>Graph</h1>
<div bind:this={container}></div>
