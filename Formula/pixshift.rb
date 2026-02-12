# Update sha256 values after each release
class Pixshift < Formula
  desc "Universal image converter CLI - convert between 10+ image formats"
  homepage "https://github.com/DanielTso/pixshift"
  version "0.1.0"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DanielTso/pixshift/releases/download/v#{version}/pixshift-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/DanielTso/pixshift/releases/download/v#{version}/pixshift-linux-amd64"
      sha256 "PLACEHOLDER_SHA256"
    end
  end

  def install
    binary_name = "pixshift-darwin-arm64" if OS.mac? && Hardware::CPU.arm?
    binary_name = "pixshift-linux-amd64" if OS.linux? && Hardware::CPU.intel?
    bin.install binary_name => "pixshift"
  end

  test do
    assert_match "pixshift", shell_output("#{bin}/pixshift --version")
  end
end
