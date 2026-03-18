import urllib.parse

class Helper(object):
    def encode_url(self, url):
      '''
      Encodes the given url that would (probably) contain characters outside the ASCII set into a valid ASCII format
      `Returns: string`
      '''
      return urllib.parse.quote(url)