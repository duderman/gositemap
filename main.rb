require 'json'
require 'open-uri'
require 'nokogiri'
require 'csv'
require 'aws-sdk-s3'
require 'ruby-progressbar'

S3_REGION    = "us-east-1"
S3_BUCKET    = "sitemap-test"
AWS_ENDPOINT = "http://127.0.0.1:9000/"

ENV['PAYLOAD'] = "{\"warehouse_locations\":[\"s3-hive-ireland\"],\"client_name\":\"tommy_hilfiger_pvh\",\"account_name\":\"tommy_hilfiger_gb_en\",\"report_name\":\"xml\",\"start_date\":\"2019-04-15\",\"provider_settings\":{\"url\":\"https://www.ferragamo.com/sfsm/sitemap_33751.xml.gz\",\"attributes_columns\":\"href,hreflang\",\"columns\":\"loc, xhtml:link, url\"}}"

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
  urls = doc.css('url')
  progressbar = ProgressBar.create(total: urls.size)
  urls.map do |url|
    progressbar.increment
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
    credentials: Aws::SharedCredentials.new(profile_name: 'minio'),
    force_path_style: true
  )

  s3.put_object(
    body: File.read(tsv_path),
    bucket: S3_BUCKET,
    key: "out.tsv"
  )
end


payload = parse_payload
url = payload['provider_settings']['url']
sitemap = download_sitemap(url)
parsed_xml = parse_sitemap(sitemap)
tsv_path = write_to_tsv(parsed_xml, url)
upload_to_s3(tsv_path)

puts 'DONE'
