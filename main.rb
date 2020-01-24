require 'json'
require 'open-uri'
require 'nokogiri'
require 'csv'
require 'aws-sdk-s3'
require 'time'

S3_BUCKET    = "sitemap-test"
S3_REGION    = ENV['S3_REGION'] || "us-east-1"
AWS_ENDPOINT = ENV['AWS_ENDPOINT'] || "http://127.0.0.1:9000/"

ENV['PAYLOAD'] ||= "{\"provider_settings\":{\"url\":\"https://www.ferragamo.com/sfsm/sitemap_33751.xml.gz\"}}"

def parse_payload
  puts 'Parsing payload'
  JSON.parse(ENV['PAYLOAD'])
end

def download_sitemap(url)
  puts 'Downloading sitemap'
  open(url).read
end

def parse_sitemap(sitemap)
  puts 'Parsing XML'
  doc = Nokogiri::XML(sitemap)
  doc.root.add_namespace('xhtml', 'http://www.w3.org/1999/xhtml')
  doc.css('url').map do |url|
    {
      'loc' => url.at_css('loc').inner_html,
      'links' => url.css("*[rel='alternate']").map do |link|
        { 'lang' => link.attribute('hreflang').value, 'href' => link.attribute('href').value }
      end
    }
  end
end

def write_to_tsv(parsed_sitemap, url)
  puts 'Writing to TSV'
  tsv = CSV.open('out.tsv', 'wb', col_sep: "\t")

  parsed_sitemap.each do |entry|
    entry['links'].each do |link|
      tsv << [entry['loc'], link['href'], link['lang'], url]
    end
  end

  tsv.flush
  tsv.close

  return tsv.path
end

def upload_to_s3(tsv_path)
  puts 'Uploading to S3'

  s3 = Aws::S3::Client.new(
    region: S3_REGION,
    endpoint: AWS_ENDPOINT,
    force_path_style: true
  )

  s3.put_object(
    body: File.read(tsv_path),
    bucket: S3_BUCKET,
    key: "out.tsv"
  )
end

startTime = Time.now
payload = parse_payload
url = payload['provider_settings']['url']
sitemap = download_sitemap(url)
parsed_xml = parse_sitemap(sitemap)
tsv_path = write_to_tsv(parsed_xml, url)
upload_to_s3(tsv_path)

puts 'DONE'
puts "It took: #{Time.now - startTime}"
puts "Memory usage: #{`ps -o rss= -p #{$$}`.to_i / 1024}"
