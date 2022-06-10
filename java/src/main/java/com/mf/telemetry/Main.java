package com.mf.telemetry;

import java.io.*;
import java.net.InetSocketAddress;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

import com.sun.net.httpserver.Headers;
import java.nio.charset.StandardCharsets;

import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapGetter;
import io.opentelemetry.context.propagation.TextMapSetter;
import io.opentelemetry.exporter.jaeger.JaegerGrpcSpanExporter;
import io.opentelemetry.exporter.jaeger.JaegerGrpcSpanExporterBuilder;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.exporter.logging.LoggingSpanExporter;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor;

import org.json.JSONArray;
import org.json.JSONObject;

import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

class JobRequest {
    String Function;
    ArrayList<Integer> Args;

    JobRequest(JSONObject o) {
        Object f = o.get("Function");
        if (f == null) {
            throw new RuntimeException("missing 'Function' key");
        }
        System.out.println("what's my class? " + f.getClass().toString() + " " + o.get("Args").getClass().toString());
        this.Function = (String) f;

        ArrayList<Integer> args = new ArrayList<>();
        JSONArray as = (JSONArray) o.get("Args");

        for (Iterator<Object> it = as.iterator(); it.hasNext(); ) {
            args.add((Integer) it.next());
        }
        this.Args = args;
    }

    JSONObject asJson() {
        JSONObject o = new JSONObject();
        o.put("Function", this.Function);
        o.put("Args", this.Args);
        return o;
    }
}

public class Main {

    public static void main(String[] args) throws IOException {
        System.out.println("starting java telemetry hack with args: " + " " + args.length + " " + new JSONArray(args).toString());

//        SdkTracerProvider sdkTracerProvider =
//                SdkTracerProvider.builder()
//                        .addSpanProcessor(JaegerGrpcSpanExporterBuilder.create(new JaegerGrpcSpanExporter()))
//                        .build();
//
//        OpenTelemetry openTelemetry = OpenTelemetrySdk.builder()
//                .setTracerProvider(sdkTracerProvider)
//                .setPropagators(ContextPropagators.create(Jae.getInstance())) // TODO B3 ?
//                .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance())) // TODO B3 ?
//                .buildAndRegisterGlobal();
//
//        Jaeg
//
//        System.out.println("opentelemetry sdk is configured: " + openTelemetry.toString());

        String bottomHost = args[0];

//        try {
//            String response = issueJsonRequest(bottomHost, "sleep", Arrays.asList(1));
//            System.out.println("response? " + response);
//        } catch (Exception e) {
//            System.out.println("OOPS!  failed to issue request: " + e.getMessage());
//        }

        // TODO need multiple threads?
        HttpServer server = HttpServer.create(new InetSocketAddress(8002), 0);
        server.createContext("/example", new JobHandler());
        addJobContext(server, bottomHost);
        server.setExecutor(null);
        server.start();
    }

    static class JobHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange t) throws IOException {
            String response = "TODO -- complete me";
            t.sendResponseHeaders(200, response.length());
            OutputStream os = t.getResponseBody();
            os.write(response.getBytes());
            os.close();
        }
    }

    static void addJobContext(HttpServer server, String bottomHost) {
        server.createContext("/job", he -> {
            try (he) {
//                OpenTelemetry openTelemetry = GlobalOpenTelemetry.get();
//                Context extractedContext = openTelemetry.getPropagators().getTextMapPropagator()
//                        .extract(Context.current(), httpExchange, getter);

                final Headers headers = he.getResponseHeaders();
                final String requestMethod = he.getRequestMethod().toUpperCase();

                String requestBody = readRequestBody(he.getRequestBody());
                JSONObject o = new JSONObject(requestBody);
                System.out.println("what did I get? " + o.toString() + " " + requestMethod + " <" + requestBody + ">");

                JobRequest jr = new JobRequest(o);

                // see: https://opentelemetry.io/docs/instrumentation/java/manual/#context-propagation
                Context extractedContext = GlobalOpenTelemetry.get().getPropagators().getTextMapPropagator()
                        .extract(Context.current(), he, getter);

                try (Scope scope = extractedContext.makeCurrent()) {

                    String bottomResponse = issueJsonRequest(bottomHost, jr.Function, jr.Args);

                    switch (requestMethod) {
                        case "POST":
                            headers.set("Content-Type", String.format("application/json; charset=%s", StandardCharsets.UTF_8));
//                        final byte[] rawResponseBody = jr.asJson().toString().getBytes(StandardCharsets.UTF_8);
                            final byte[] rawResponseBody = bottomResponse.getBytes(StandardCharsets.UTF_8);
                            he.sendResponseHeaders(200, rawResponseBody.length);
                            he.getResponseBody().write(rawResponseBody);
                            break;
                        default:
                            headers.set("Allow", "POST");
                            he.sendResponseHeaders(405, -1);
                            break;
                    }
                }
            }
        });
    }

    public static String readRequestBody(InputStream is) throws IOException {
        InputStreamReader isr =  new InputStreamReader(is,"utf-8");
        BufferedReader br = new BufferedReader(isr);
        String text = "";
        while (true) {
            String value = br.readLine();
            if (value == null) {
                break;
            }
            text += value;
        }
        return text;
    }

    public static String issueJsonRequest(String host, String name, List<Integer> args) throws IOException {
        JSONObject data = new JSONObject();
        data.put("name", name);
        data.put("args", args);

        URL url = new URL("http://" + host + ":8003/function");
        HttpURLConnection httpConnection  = (HttpURLConnection) url.openConnection();
        httpConnection.setDoOutput(true);
        httpConnection.setRequestMethod("POST");
        httpConnection.setRequestProperty("Content-Type", "application/json");
        httpConnection.setRequestProperty("Accept", "application/json");

        // see: https://opentelemetry.io/docs/instrumentation/java/manual/#context-propagation
        GlobalOpenTelemetry.get().getPropagators().getTextMapPropagator().inject(Context.current(), httpConnection, setter);

        System.out.println("encoded? " + data.toString());

        DataOutputStream wr = new DataOutputStream(httpConnection.getOutputStream());
        wr.write(data.toString().getBytes());
        int responseCode = httpConnection.getResponseCode();

        BufferedReader bufferedReader;
        if (responseCode > 199 && responseCode < 300) {
            bufferedReader = new BufferedReader(new InputStreamReader(httpConnection.getInputStream()));
        } else {
            bufferedReader = new BufferedReader(new InputStreamReader(httpConnection.getErrorStream()));
        }

        StringBuilder content = new StringBuilder();
        String line;
        while ((line = bufferedReader.readLine()) != null) {
            content.append(line).append("\n");
        }
        bufferedReader.close();

        return content.toString();
    }

    static TextMapSetter<HttpURLConnection> setter =
        new TextMapSetter<HttpURLConnection>() {
            @Override
            public void set(HttpURLConnection carrier, String key, String value) {
                // Insert the context as Header
                carrier.setRequestProperty(key, value);
            }
        };

    static TextMapGetter<HttpExchange> getter =
        new TextMapGetter<>() {
            @Override
            public String get(HttpExchange carrier, String key) {
                if (carrier.getRequestHeaders().containsKey(key)) {
                    return carrier.getRequestHeaders().get(key).get(0);
                }
                return null;
            }

            @Override
            public Iterable<String> keys(HttpExchange carrier) {
                return carrier.getRequestHeaders().keySet();
            }
        };

}
