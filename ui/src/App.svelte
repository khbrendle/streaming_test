<script>
		import vis from "vis-timeline/dist/vis-timeline-graph2d.min";
		import { onMount, onDestroy } from "svelte";
		import Datepicker from "../node_modules/svelte-calendar/src/Components/Datepicker.svelte";

		let stream = "customer_count";
		let response = "n";
		let container;
		let graph2d;

		let streamOptions = {
		  customer_count: [{ value: "n", display: "Count" }],
		  order_count: [
		    { value: "n", display: "Count" },
		    { value: "revenue", display: "Revenue" }
		  ]
		};

		var dataset = new vis.DataSet([]);
		// let dataView = new vis.DataView(dataset, {
		//   fields: {
		//     x: "x",
		//     n: "y",
		//     y: "n"
		//   }
		// });

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
		// start date for history
		var startDate;
		var startDateSelected;

		let groupMinuteOptions = [1, 2, 3, 4, 5, 10, 15, 30];

		const handleChangeGroupMinute = () => {
		  console.log("closing current stream");
		  dataStream.close();
		  dataset.clear();
		  streamHistory = true;

		  if (startDateSelected) {
		    console.log("selected startDate: ", startDate);
		  }

		  // dataView = new vis.DataView(dataset, {
		  //   fields: {
		  //     x: "x",
		  //     y: "n"
		  //   }
		  // });

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
		    console.log("got :", dat);
		    // new implementation with every data of array

		    dat.map(v => {
		      var posX = new Date(v.time_stamp);
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

		      if (val.length === 0) {
		        // does not exist, create
		        console.log("adding new value");
		        // val = {
		        //   x: posXn,
		        //   y: dat.n
		        // };

		        v.x = posXn;
		        v.y = v[response];
		        console.log("new value: ", v);
		        dataset.add(v);
		      } else {
		        val = val[0];
		        // exists, update
		        // console.log("updating existing value");
		        console.log("current value: ", val);
		        let keys = Object.keys(val);
		        let key;
		        for (var i = 0; i < keys.length; i++) {
		          key = keys[i];
		          // if (key === "x" || key === "id" || key === "timestamp") {
		          if (key != "n") {
		            // testing with just 1 response
		            continue;
		          }
		          console.log("key: ", key);
		          console.log("was: ", val.y);
		          val.y = val.y + v[key];
		          console.log("is: ", val.y);
		        }
		        // val.y += dat.n;
		        console.log("new value: ", val);
		        dataset.update(val);
		      }
		    });

		    // // old implementation with array of history and single object events
		    // // this method is more efficient handling the historic data because it can do
		    // // a bulk insert to the data but requires this custom handling of first event
		    // if (streamHistory) {
		    //   dat = dat.map(v => {
		    //     // seconds already set to 0
		    //     v.x = new Date(v.x).setMilliseconds(0);
		    //     v.y = v[response];
		    //     return v;
		    //   });
		    //   console.log("setting: ", dat);
		    //   dataset.add(dat);
		    //   streamHistory = false;
		    //   return;
		    // }
		    //
		    // var posX = new Date(dat.timestamp);
		    // posX = new Date(new Date(posX.setSeconds(0)).setMilliseconds(0));
		    // var min = posX.getMinutes();
		    // min = Math.floor(min / groupMinute) * groupMinute;
		    // var posXn = posX.setMinutes(min);
		    //
		    // var val = dataset.get(
		    //   {
		    //     filter: function(item) {
		    //       // console.log("filter item: ", item);
		    //       return item.x == posXn;
		    //     }
		    //   },
		    //   {
		    //     type: {
		    //       y: "Number"
		    //     }
		    //   }
		    // );
		    // // console.log("filter got: ", val);
		    // if (val.length === 0) {
		    //   // does not exist, create
		    //   console.log("adding new value");
		    //   // val = {
		    //   //   x: posXn,
		    //   //   y: dat.n
		    //   // };
		    //
		    //   dat.x = posXn;
		    //   dat.y = dat[response];
		    //   // console.log("new value: ", val);
		    //   dataset.add(dat);
		    // } else {
		    //   val = val[0];
		    //   // exists, update
		    //   // console.log("updating existing value");
		    //   console.log("current value: ", val);
		    //   let keys = Object.keys(val);
		    //   let key;
		    //   for (var i = 0; i < keys.length; i++) {
		    //     key = keys[i];
		    //     // if (key === "x" || key === "id" || key === "timestamp") {
		    //     if (key != "n") {
		    //       // testing with just 1 response
		    //       continue;
		    //     }
		    //     console.log("key: ", key);
		    //     console.log("was: ", val[key]);
		    //     val[key] = val[key] + dat[key];
		    //     console.log("is: ", val[key]);
		    //   }
		    //   // val.y += dat.n;
		    //   console.log("new value: ", val);
		    //   dataset.update(val);
		    // }
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
			<label for="stream-response-select">Response</label>
			<select id="stream-response-select" class="form-control form-control-sm" bind:value={response}>
				{#each streamOptions[stream] as o}
				<option value={o.value}>{o.display}</option>
				{/each}
			</select>
		</div>

		<div class="form-row">
			<div class="form-group col-md-6">
				<label for="customer-stream-group-min">Minute Granularity</label>
				<select id="customer-stream-group-min" class="form-control form-control-sm" bind:value={groupMinute}>
					{#each groupMinuteOptions as o}
						<option value={o}>{o}</option>
					{/each}
				</select>
			</div>

			<div class="form-group col-md-6">
				<label for="start-date-picker">Start Date</label>
				<div id="start-date-picker">
					<Datepicker bind:selected={startDate} bind:dateChosen={startDateSelected}>
						<button type="button" class="btn btn-sm btn-secondary">Pick Date</button>
					</Datepicker>
				</div>
			</div>
		</div>

		<button type="button" class="btn btn-primary" on:click={handleChangeGroupMinute}>Submit</button>
		<button type="button" class="btn btn-danger" on:click={() => {dataStream.close()}}>Stop</button>
		<button type="button" class="btn btn-info" on:click={() => {console.log("vis data: ",dataset.get())}}>Log Data</button>
	</div>
</div>
<div bind:this={container}></div>
