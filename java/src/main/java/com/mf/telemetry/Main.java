package com.mf.telemetry;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;

import com.google.gson.JsonArray;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

import com.sun.net.httpserver.Headers;
import java.nio.charset.StandardCharsets;

import org.json.JSONObject;
import java.io.BufferedReader;
import java.io.DataOutputStream;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;


public class Main {

    public static void main(String[] args) throws IOException {
        System.out.println("starting java telemetry hack");

        try {
            String response = issueJsonRequest("localhost", "sleep", new int[]{1});
            System.out.println("response? " + response);
        } catch (Exception e) {
            System.out.println("OOPS!  failed to issue request: " + e.getMessage());
        }

        HttpServer server = HttpServer.create(new InetSocketAddress(8004), 0);
        server.createContext("/example", new JobHandler());
        addJobContext(server);
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

    static void addJobContext(HttpServer server) throws IOException {
        server.createContext("/job", he -> {
            try {
                final Headers headers = he.getResponseHeaders();
                final String requestMethod = he.getRequestMethod().toUpperCase();
                switch (requestMethod) {
                    case "POST":
                        final String responseBody = "{\"abc\": 123}";
                        headers.set("Content-Type", String.format("application/json; charset=%s", StandardCharsets.UTF_8));
                        final byte[] rawResponseBody = responseBody.getBytes(StandardCharsets.UTF_8);
                        he.sendResponseHeaders(200, rawResponseBody.length);
                        he.getResponseBody().write(rawResponseBody);
                        break;
                    default:
                        headers.set("Allow", "POST");
                        he.sendResponseHeaders(405, -1);
                        break;
                }
            } finally {
                he.close();
            }
        });
    }

    public static String issueJsonRequest(String host, String name, int[] args) throws IOException {
        JSONObject data = new JSONObject();
        data.put("name", name);
        data.put("args", args);

        URL url = new URL("http://" + host + ":8003/function");
        HttpURLConnection httpConnection  = (HttpURLConnection) url.openConnection();
        httpConnection.setDoOutput(true);
        httpConnection.setRequestMethod("POST");
        httpConnection.setRequestProperty("Content-Type", "application/json");
        httpConnection.setRequestProperty("Accept", "application/json");

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

}
