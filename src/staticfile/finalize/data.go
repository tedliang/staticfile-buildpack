package finalize

const (
	initScript = `
# ------------------------------------------------------------------------------------------------
# Copyright 2013 Jordon Bedwell.
# Apache License.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
# except  in compliance with the License. You may obtain a copy of the License at:
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the
# License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the specific language governing permissions
# and  limitations under the License.
# ------------------------------------------------------------------------------------------------

export APP_ROOT=$HOME
export LD_LIBRARY_PATH=$APP_ROOT/nginx/lib:$LD_LIBRARY_PATH

mv $APP_ROOT/nginx/conf/nginx.conf $APP_ROOT/nginx/conf/nginx.conf.erb
erb $APP_ROOT/nginx/conf/nginx.conf.erb > $APP_ROOT/nginx/conf/nginx.conf

if [[ ! -f $APP_ROOT/nginx/logs/access.log ]]; then
    mkfifo $APP_ROOT/nginx/logs/access.log
fi

if [[ ! -f $APP_ROOT/nginx/logs/error.log ]]; then
    mkfifo $APP_ROOT/nginx/logs/error.log
fi
`

	startLoggingScript = `
cat < $APP_ROOT/nginx/logs/access.log &
(>&2 cat) < $APP_ROOT/nginx/logs/error.log &
`

	startCommand = `#!/bin/sh
set -ex
$APP_ROOT/start_logging.sh
nginx -p $APP_ROOT/nginx -c $APP_ROOT/nginx/conf/nginx.conf
`

	nginxConfTemplate = `
worker_processes 1;
daemon off;

error_log <%= ENV["APP_ROOT"] %>/nginx/logs/error.log;
events { worker_connections 1024; }

http {
  charset utf-8;
  log_format cloudfoundry '$http_x_forwarded_for - $http_referer - [$time_local] "$request" $status $body_bytes_sent';
  access_log <%= ENV["APP_ROOT"] %>/nginx/logs/access.log cloudfoundry;
  default_type application/octet-stream;
  include mime.types;
  sendfile on;

  gzip on;
  gzip_disable "msie6";
  gzip_comp_level 6;
  gzip_min_length 1100;
  gzip_buffers 16 8k;
  gzip_proxied any;
  gunzip on;
  gzip_static always;
  gzip_types text/plain text/css text/js text/xml text/javascript application/javascript application/x-javascript application/json application/xml application/xml+rss;
  gzip_vary on;

  tcp_nopush on;
  keepalive_timeout 30;
  port_in_redirect off; # Ensure that redirects don't include the internal container PORT - <%= ENV["PORT"] %>
  server_tokens off;

  server {
    listen <%= ENV["PORT"] %>;
    server_name localhost;

    root <%= ENV["APP_ROOT"] %>/public;
    index  index.html;

    {{if .ForceHTTPS}}
      set $updated_host $host;
      if ($http_x_forwarded_host != "") {
        set $updated_host $http_x_forwarded_host;
      } 

      if ($http_x_forwarded_proto != "https") {
        return 301 https://$updated_host$request_uri;
      }
    {{else}}
      <% if ENV["FORCE_HTTPS"] %>
        set $updated_host $host;
        if ($http_x_forwarded_host != "") {
          set $updated_host $http_x_forwarded_host;
        } 

        if ($http_x_forwarded_proto != "https") {
          return 301 https://$updated_host$request_uri;
        }
      <% end %>
    {{end}}


    location / {  
      <% if ENV["PRERENDER_TOKEN"] && ENV["PRERENDER_TOKEN"].length > 0 %>
        try_files $uri @prerender;
      <% else %>
        {{if .PushState}}
          if (!-e $request_filename) {
            rewrite ^(.*)$ / break;
          }
        {{end}}
      <% end %>

      {{if .DirectoryIndex}}
        autoindex on;
      {{end}}

      {{if .BasicAuth}}
        auth_basic "Restricted";  #For Basic Auth
        auth_basic_user_file <%= ENV["APP_ROOT"] %>/nginx/conf/.htpasswd;
      {{end}}

      {{if .SSI}}
        ssi on;
      {{end}}

      {{if .HSTS}}
        add_header Strict-Transport-Security "max-age=31536000{{if .HSTSIncludeSubDomains}}; includeSubDomains{{end}}{{if .HSTSPreload}}; preload{{end}}";
      {{end}}

      {{if ne .LocationInclude ""}}
        include {{.LocationInclude}};
      {{end}}

			{{ range $code, $value := .StatusCodes }}
			  error_page {{ $code }} {{ $value }};
		  {{ end }}
    }

    <% if ENV["PRERENDER_TOKEN"] && ENV["PRERENDER_TOKEN"].length > 0 %>
    location @prerender {
          proxy_set_header X-Prerender-Token <%= ENV["PRERENDER_TOKEN"] %>;
	  
	  set $prerender 0;
          if ($http_user_agent ~* "googlebot|bingbot|yandex|baiduspider|Screaming Frog SEO Spider|twitterbot|facebookexternalhit|rogerbot|linkedinbot|embedly|quora link preview|showyoubot|outbrain|pinterest\/0\.|pinterestbot|slackbot|vkShare|W3C_Validator|whatsapp") {
              set $prerender 1;
          }
          if ($args ~ "_escaped_fragment_") {
              set $prerender 1;
          }
          if ($http_user_agent ~ "Prerender") {
              set $prerender 0;
          }
          if ($uri ~* "\.(js|css|xml|less|png|jpg|jpeg|gif|pdf|doc|txt|ico|rss|zip|mp3|rar|exe|wmv|doc|avi|ppt|mpg|mpeg|tif|wav|mov|psd|ai|xls|mp4|m4a|swf|dat|dmg|iso|flv|m4v|torrent|ttf|woff|svg|eot)") {
              set $prerender 0;
          }

          #resolve using Google's DNS server to force DNS resolution and prevent caching of IPs
          resolver 8.8.8.8;

          if ($prerender = 1) {
              #setting prerender as a variable forces DNS resolution since nginx caches IPs and doesnt play well with load balancing
              set $prerender "service.prerender.io";
	      rewrite .* /https://$host$request_uri? break;
              proxy_pass http://$prerender;
          }

          if ($prerender = 0) {
              rewrite .* /index.html break;
	  }
    }
    <% end %>

    {{if not .HostDotFiles}}
      location ~ /\. {
        deny all;
        return 404;
      }
    {{end}}
  }
}
`
	MimeTypes = `
types {
  text/html html htm shtml;
  text/css css;
  text/xml xml;
  image/gif gif;
  image/jpeg jpeg jpg;
  application/javascript js;
  application/atom+xml atom;
  application/rss+xml rss;
  font/ttf ttf;
  font/woff woff;
  font/woff2 woff2;
  text/mathml mml;
  text/plain txt;
  text/vnd.sun.j2me.app-descriptor jad;
  text/vnd.wap.wml wml;
  text/x-component htc;
  text/cache-manifest manifest;
  image/png png;
  image/tiff tif tiff;
  image/vnd.wap.wbmp wbmp;
  image/x-icon ico;
  image/x-jng jng;
  image/x-ms-bmp bmp;
  image/svg+xml svg svgz;
  image/webp webp;
  application/java-archive jar war ear;
  application/mac-binhex40 hqx;
  application/msword doc;
  application/pdf pdf;
  application/postscript ps eps ai;
  application/rtf rtf;
  application/vnd.ms-excel xls;
  application/vnd.ms-powerpoint ppt;
  application/vnd.wap.wmlc wmlc;
  application/vnd.google-earth.kml+xml  kml;
  application/vnd.google-earth.kmz kmz;
  application/x-7z-compressed 7z;
  application/x-cocoa cco;
  application/x-java-archive-diff jardiff;
  application/x-java-jnlp-file jnlp;
  application/x-makeself run;
  application/x-perl pl pm;
  application/x-pilot prc pdb;
  application/x-rar-compressed rar;
  application/x-redhat-package-manager  rpm;
  application/x-sea sea;
  application/x-shockwave-flash swf;
  application/x-stuffit sit;
  application/x-tcl tcl tk;
  application/x-x509-ca-cert der pem crt;
  application/x-xpinstall xpi;
  application/xhtml+xml xhtml;
  application/zip zip;
  application/octet-stream bin exe dll;
  application/octet-stream deb;
  application/octet-stream dmg;
  application/octet-stream eot;
  application/octet-stream iso img;
  application/octet-stream msi msp msm;
  application/json json;
  audio/midi mid midi kar;
  audio/mpeg mp3;
  audio/ogg ogg;
  audio/x-m4a m4a;
  audio/x-realaudio ra;
  video/3gpp 3gpp 3gp;
  video/mp4 mp4;
  video/mpeg mpeg mpg;
  video/quicktime mov;
  video/webm webm;
  video/x-flv flv;
  video/x-m4v m4v;
  video/x-mng mng;
  video/x-ms-asf asx asf;
  video/x-ms-wmv wmv;
  video/x-msvideo avi;
}
`
)
