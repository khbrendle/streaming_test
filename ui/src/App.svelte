<script>
	import vis from "vis-timeline/dist/vis-timeline-graph2d.min";
	import { onMount, onDestroy } from "svelte";

	let stream = "customer_count";
	let container;
	let graph2d;

	var dataset = new vis.DataSet([]);

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
	  },
	  // timeAxis: { scale: "minute" },
	  graphHeight: "300px"
	};

	let dataStream;
	let streamHistory = true;
	// default 1 minute
	let groupMinute;

	let groupMinuteOptions = [1, 2, 3, 4, 5, 10, 15, 30];

	const handleChangeGroupMinute = () => {
	  console.log("closing current stream");
	  dataStream.close();
	  dataset.clear();
	  streamHistory = true;

	  dataStream = connectStream(stream, {
	    min: groupMinute
	  });
	};

	function connectStream(stream, options) {
	  // console.log("options: ", options);
	  // initialize query params
	  var qp = "";
	  if (Object.keys(options).includes("min")) {
	    qp += `?groupMinute=${options.min}`;
	  } else {
	    throw new Error("missing required option `groupMinute`");
	  }
	  console.log(
	    `connecting to stream ${stream} with ${options.min} minute granularity`
	  );
	  var s = new EventSource(
	    `http://localhost:3000/v0/stream/subscribe/${stream}${qp}`
	  );

	  s.onmessage = function(event) {
	    var dat = JSON.parse(event.data);

	    if (streamHistory) {
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
	  s.onerror = function(err) {
	    console.log("stream error: ", err);
	  };

	  return s;
	}

	onMount(() => {
	  dataStream = connectStream(stream, {
	    min: groupMinute
	  });

	  graph2d = new vis.Graph2d(container, dataset, options);
	});

	onDestroy(() => {
	  dataStream.close();
	});
</script>
<svelte:head>
	<link rel="stylesheet" href="/vis-timeline-graph2d.min.css">
</svelte:head>

<h1>Live View</h1>
<div class="card">
	<div class="card-body">
		<div class="form-group">
			<label for="stream-name-select">Stream Name</label>
			<select id="stream-name-select" class="form-control form-control-sm" bind:value={stream}>
				<option value="customer_count">Customer</option>
				<option value="order_count">Order</option>
			</select>
		</div>

		<div class="form-group">
			<label for="customer-stream-group-min">Minute Granularity</label>
			<select id="customer-stream-group-min" class="form-control form-control-sm" bind:value={groupMinute}>
				{#each groupMinuteOptions as o}
					<option value={o}>{o}</option>
				{/each}
			</select>
		</div>

		<button type="button" class="btn btn-primary" on:click={handleChangeGroupMinute}>Submit</button>
		<button type="button" class="btn btn-danger" on:click={() => {dataStream.close()}}>Stop</button>
	</div>
</div>
<div bind:this={container}></div>
